package api

import "github.com/leobishop234/util/srverr"

func ParseHttpError(statusCode int, msgPrefix string, err error) error {
	switch {
	case statusCode == 503:
		return srverr.New(srverr.ErrCodeDependencyFailure, msgPrefix, err)
	case statusCode >= 500:
		return srverr.New(srverr.ErrCodeInternal, msgPrefix, err)
	case statusCode == 429:
		return srverr.New(srverr.ErrCodeRateLimited, msgPrefix, err)
	case statusCode == 410:
		return srverr.New(srverr.ErrCodePurged, msgPrefix, err)
	case statusCode == 409:
		return srverr.New(srverr.ErrCodeStateConflict, msgPrefix, err)
	case statusCode == 408:
		return srverr.New(srverr.ErrCodeTimeout, msgPrefix, err)
	case statusCode == 412:
		return srverr.New(srverr.ErrCodePreconditionFailed, msgPrefix, err)
	case statusCode == 403:
		return srverr.New(srverr.ErrCodeAuthorizationFailure, msgPrefix, err)
	case statusCode == 401:
		return srverr.New(srverr.ErrCodeAuthenticationFailure, msgPrefix, err)
	case statusCode == 404:
		return srverr.New(srverr.ErrCodeNotFound, msgPrefix, err)
	case statusCode >= 400:
		return srverr.New(srverr.ErrCodeValidation, msgPrefix, err)
	default:
		return srverr.New(srverr.ErrCodeInternal, msgPrefix, err)
	}
}
