package model

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrBadRequest    = errors.New("bad request")
	ErrWrongPassword = errors.New("wrong credentials")
)

var tagMsg = map[string]string{
	"required": "required field",
	"min":      "minimum %s character(s)",
	"max":      "maximum %s character(s)",
	"len":      "exactly %s character(s)",
}

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }

type PublicError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func ToHTTP(err error) (int, PublicError) {
	var v *ValidationError
	if errors.As(err, &v) {
		return http.StatusUnprocessableEntity, PublicError{
			Code: "validation_failed", Message: "Please check field values", Details: v.Fields,
		}
	}
	switch {
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, PublicError{Code: "unauthorized", Message: "Authorization required"}
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, PublicError{Code: "forbidden", Message: "Access denied"}
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, PublicError{Code: "not_found", Message: "Resource not found"}
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, PublicError{Code: "conflict", Message: "State conflict"}
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, PublicError{Code: "bad_request", Message: "Bad request"}
	case errors.Is(err, ErrWrongPassword):
		return http.StatusUnauthorized, PublicError{Code: "wrong_password", Message: "Invalid password"}
	default:
		return http.StatusInternalServerError, PublicError{
			Code: "internal_error", Message: "Internal server error",
		}
	}
}

func AsValidationError(req any, err error) (*ValidationError, bool) {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return nil, false
	}
	fields := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		field := jsonFieldName(req, fe.StructField())
		tmpl, ok := tagMsg[fe.Tag()]
		var msg string
		if ok {
			if strings.Contains(tmpl, "%s") {
				msg = fmt.Sprintf(tmpl, fe.Param())
			} else {
				msg = tmpl
			}
		} else {
			msg = fe.Error()
		}
		fields[field] = msg
	}
	return &ValidationError{Fields: fields}, true
}

func jsonFieldName(obj any, structField string) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if f, ok := t.FieldByName(structField); ok {
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			return strings.ToLower(structField)
		}
		if i := strings.Index(tag, ","); i > 0 {
			return tag[:i]
		}
		return tag
	}
	return strings.ToLower(structField)
}
