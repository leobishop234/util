package cleanup

import (
	"errors"
	"io"
	"testing"

	"github.com/leobishop234/util/srverr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCloser struct {
	closeErr error
	closed   bool
}

func (tc *testCloser) Close() error {
	tc.closed = true
	return tc.closeErr
}

func newTestCloser(closeErr error) *testCloser {
	return &testCloser{closeErr: closeErr}
}

func TestClose(t *testing.T) {
	t.Parallel()

	closeFailedErr := errors.New("close failed") //nolint:err113 // test error

	tests := map[string]struct {
		closer      io.Closer
		resource    string
		wantClosed  bool
		wantErrCode srverr.ErrCode
		wantErrMsg  string
		wantWrapped error
		expectErr   bool
	}{
		"returns nil when closer is nil": {
			closer:    nil,
			resource:  "test resource",
			expectErr: false,
		},
		"returns nil when close succeeds": {
			closer:     newTestCloser(nil),
			resource:   "test resource",
			wantClosed: true,
			expectErr:  false,
		},
		"wraps close failures as internal srverr": {
			closer:      newTestCloser(closeFailedErr),
			resource:    "test resource",
			wantClosed:  true,
			wantErrCode: srverr.ErrCodeInternal,
			wantErrMsg:  "failed to close resource: test resource",
			wantWrapped: closeFailedErr,
			expectErr:   true,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			err := Close(testCase.closer, testCase.resource)
			if !testCase.expectErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				srvErr, ok := srverr.Unwrap(err)
				require.True(t, ok)
				assert.Equal(t, testCase.wantErrCode, srvErr.Code())
				assert.Equal(t, testCase.wantErrMsg, srvErr.Message())
				require.ErrorIs(t, srvErr.Err(), testCase.wantWrapped)
			}

			if testCloser, ok := testCase.closer.(*testCloser); ok {
				assert.Equal(t, testCase.wantClosed, testCloser.closed)
			}
		})
	}
}
