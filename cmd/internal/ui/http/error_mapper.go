package http

import (
	"errors"
	"net/http"

	apperror "github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/error"
)

type InternalServerError struct{}

func NewInternalServerError() *InternalServerError {
	return &InternalServerError{}
}

func (e *InternalServerError) Error() string {
	return "INTERNAL_SERVER_ERROR"
}

func MapError(err error) (int, error) {
	var clientAlreadyExists *apperror.ClientAlreadyExists
	var clientNotFound *apperror.ClientNotFound

	switch {
	case errors.As(err, &clientAlreadyExists):
		return http.StatusUnprocessableEntity, err
	case errors.As(err, &clientNotFound):
		return http.StatusNotFound, err
	default:
		return http.StatusInternalServerError, NewInternalServerError()
	}
}
