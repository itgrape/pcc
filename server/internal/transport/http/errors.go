package httpapi

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/itgrape/pcc/server/internal/domain"
)

func toHTTPError(err error) error {
	if err == nil {
		return nil
	}
	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		return huma.NewError(http.StatusInternalServerError, "internal error")
	}
	return huma.NewError(statusFor(domainErr.Kind), domainErr.Message)
}

func statusFor(kind domain.Kind) int {
	switch kind {
	case domain.KindValidation:
		return http.StatusBadRequest
	case domain.KindUnauthed:
		return http.StatusUnauthorized
	case domain.KindForbidden:
		return http.StatusForbidden
	case domain.KindNotFound:
		return http.StatusNotFound
	case domain.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
