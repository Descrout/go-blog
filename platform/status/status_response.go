package status

import (
	"net/http"

	"github.com/go-chi/render"
)

type StatusResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *StatusResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &StatusResponse{
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrInternal(err error) render.Renderer {
	return &StatusResponse{
		HTTPStatusCode: 500,
		StatusText:     "Internal Server Error",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &StatusResponse{
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func ErrConflict(err string) render.Renderer {
	return &StatusResponse{
		HTTPStatusCode: 409,
		StatusText:     "Conflict.",
		ErrorText:      err,
	}
}

func ErrUnauthorized(err string) render.Renderer {
	return &StatusResponse{
		HTTPStatusCode: 401,
		StatusText:     "Unauthorized.",
		ErrorText:      err,
	}
}

func DelSuccess() render.Renderer {
	return &StatusResponse{
		HTTPStatusCode: 200,
		StatusText:     "Successfuly deleted.",
	}
}

var ErrNotFound = &StatusResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
