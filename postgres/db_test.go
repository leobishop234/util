package postgres

import (
	"testing"
)

func TestBuildDBURL(t *testing.T) {
	base := Config{
		Host:     "localhost",
		Port:     5432,
		Database: "mydb",
		User:     "alice",
		Pass:     "secret",
		Sslmode:  "require",
	}

	t.Run("no options", func(t *testing.T) {
		got := buildDBURL(base)
		want := "postgres://alice:secret@localhost:5432/mydb?sslmode=require"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("single option", func(t *testing.T) {
		c := base
		c.Options = map[string]string{"databaseid": "abc123"}
		got := buildDBURL(c)
		want := "postgres://alice:secret@localhost:5432/mydb?sslmode=require&options=databaseid%3Dabc123"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("multiple options sorted deterministically", func(t *testing.T) {
		c := base
		c.Options = map[string]string{
			"search_path": "public",
			"databaseid":  "abc123",
		}
		got := buildDBURL(c)
		// Keys are sorted: databaseid before search_path
		want := "postgres://alice:secret@localhost:5432/mydb?sslmode=require&options=databaseid%3Dabc123+search_path%3Dpublic"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
