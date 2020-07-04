package errors

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrInternal(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: 500,
		StatusText:     "Internal Server Error",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func ErrConflict(err string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: 409,
		StatusText:     "Conflict.",
		ErrorText:      err,
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
var ErrUnauthorized = &ErrResponse{HTTPStatusCode: 401, StatusText: "Unauthorized."}
