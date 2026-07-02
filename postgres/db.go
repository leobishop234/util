package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ariga.io/atlas/atlasexec"
	"github.com/jmoiron/sqlx"
	"github.com/leobishop234/util/cleanup"
	"github.com/leobishop234/util/srverr"
)

const (
	PsqlMaxParams = 65535
)

type DBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Pass     string //nolint:gosec // internal config not exposed publicly
	Sslmode  string
	Options  map[string]string
}

type MigrationConfig struct {
	MigrationsEnabled bool
	AtlasBin          string
	DevDB             DBConfig
}

type DB struct {
	*sqlx.DB
	config DBConfig
}

// Option configures InitDB behaviour.
type Option func(*dbOptions)

type dbOptions struct {
	migration *migrationOpts
}

type migrationOpts struct {
	conf   MigrationConfig
	schema embed.FS
}

// Migration returns an Option that runs Atlas schema migrations against conf
// using devDB as the Atlas dev database and schema as the embedded SQL files.
func Migration(conf MigrationConfig, schema embed.FS) Option {
	return func(o *dbOptions) {
		o.migration = &migrationOpts{conf: conf, schema: schema}
	}
}

func InitDB(ctx context.Context, conf DBConfig, opts ...Option) (*DB, error) {
	o := &dbOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if o.migration != nil && o.migration.conf.MigrationsEnabled {
		schDir, clnup, err := schemaDir(o.migration.schema)
		if err != nil {
			return nil, err
		}
		defer clnup()

		if err := applySchema(ctx, conf, o.migration.conf.DevDB, schDir, o.migration.conf.atlasBin()); err != nil {
			return nil, err
		}
	}

	db, err := openDB(ctx, conf)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	if err := db.DB.Close(); err != nil {
		return Error("failed to close database", err, srverr.ErrCodeDependencyFailure)
	}

	return nil
}

func openDB(ctx context.Context, conf DBConfig) (*DB, error) {
	db, err := sqlx.Open(string("postgres"), buildDBURL(conf))
	if err != nil {
		return nil, Error("failed to open database", err, srverr.ErrCodeDependencyFailure)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, Error("failed to ping database", err, srverr.ErrCodeDependencyFailure)
	}

	return &DB{DB: db, config: conf}, nil
}

func (c MigrationConfig) atlasBin() string {
	if c.AtlasBin == "" {
		return "atlas"
	}

	return c.AtlasBin
}

func applySchema(ctx context.Context, conf, devConf DBConfig, schemaDir, atlasBin string) (err error) {
	workDir, err := atlasexec.NewWorkingDir()
	if err != nil {
		return Error("failed to create atlas working directory", err, srverr.ErrCodeInternal)
	}
	defer func() {
		err = errors.Join(err, cleanup.Close(workDir, "atlas working directory"))
	}()

	client, err := atlasexec.NewClient(workDir.Path(), atlasBin)
	if err != nil {
		return Error("failed to create atlas client", err, srverr.ErrCodeDependencyFailure)
	}

	// Apply the schema using Atlas
	_, err = client.SchemaApply(ctx, &atlasexec.SchemaApplyParams{
		URL:         buildDBURL(conf),
		To:          "file://" + schemaDir,
		DevURL:      buildDBURL(devConf),
		AutoApprove: true,
	})
	if err != nil {
		return Error("failed to apply schema", err, srverr.ErrCodeDependencyFailure)
	}

	return nil
}

func buildDBURL(aConfig DBConfig) string {
	u := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		aConfig.User, aConfig.Pass, aConfig.Host, aConfig.Port, aConfig.Database, aConfig.Sslmode)

	if len(aConfig.Options) > 0 {
		keys := make([]string, 0, len(aConfig.Options))
		for k := range aConfig.Options {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+"="+aConfig.Options[k])
		}

		u += "&options=" + url.QueryEscape(strings.Join(parts, " "))
	}

	return u
}

func schemaDir(schemaFS embed.FS) (tmpDir string, clnup func(), err error) {
	tmpDir, err = os.MkdirTemp("", "book-my-holiday-schema-")
	if err != nil {
		return "", nil, Error("failed to create temporary schema directory", err, srverr.ErrCodeInternal)
	}

	// Walk the entire schema directory recursively
	err = fs.WalkDir(schemaFS, "schema", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return Error("failed to walk embedded schema directory", err, srverr.ErrCodeInternal)
		}

		// Skip the root "schema" directory itself
		if path == "schema" {
			return nil
		}

		// Calculate the relative path from "schema"
		relPath, err := filepath.Rel("schema", path)
		if err != nil {
			return Error("failed to calculate relative path", err, srverr.ErrCodeInternal)
		}

		destPath := filepath.Join(tmpDir, relPath)

		if d.IsDir() {
			// Create subdirectory
			if err := os.MkdirAll(destPath, 0o750); err != nil {
				return Error("failed to create subdirectory", err, srverr.ErrCodeInternal)
			}
			return nil
		}

		// Copy file
		content, err := fs.ReadFile(schemaFS, path)
		if err != nil {
			return Error("failed to read embedded schema file", err, srverr.ErrCodeInternal)
		}

		if err := os.WriteFile(destPath, content, 0o600); err != nil {
			return Error("failed to write schema file", err, srverr.ErrCodeInternal)
		}

		return nil
	})
	if err != nil {
		return "", nil, err
	}

	return tmpDir, func() {
		_ = os.RemoveAll(tmpDir)
	}, nil
}
