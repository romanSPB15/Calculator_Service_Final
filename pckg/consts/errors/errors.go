package errors

import (
	"io"
	"net/http"
)

const (
	InvalidBody         = "invalid body"
	UserAlreadyExists   = "user already exists"
	ShortPassword       = "short password"
	UserNotFound        = "user not found"
	InternalServerError = "internal server error"
)

func ResponseError(resp *http.Response) string {
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(b)
}
