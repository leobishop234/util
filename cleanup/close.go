package cleanup

import (
	"io"

	"github.com/leobishop234/util/srverr"
)

// Close closes closer and returns a wrapped error if closing fails.
func Close(closer io.Closer, resource string) error {
	if closer == nil {
		return nil
	}

	if err := closer.Close(); err != nil {
		return srverr.New(srverr.ErrCodeInternal, "failed to close resource: "+resource, err)
	}

	return nil
}
