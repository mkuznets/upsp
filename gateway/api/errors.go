package api

import (
	"github.com/go-chi/render"
	"log"
	"net/http"
)

// Error represents an HTTP error returned from the Api.
type Error struct {
	Err  error
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

// Json returns a JSON representation of the error.
func (e *Error) Json() render.M {
	return render.M{
		"error":   http.StatusText(e.Code),
		"message": e.Msg,
	}
}

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	switch v := err.(type) {
	case *Error:
		render.Status(r, v.Code)
		render.JSON(w, r, v.Json())
	default:
		log.Printf("[ERR] %v", err)
		e := &Error{Err: err, Code: http.StatusInternalServerError, Msg: "Unexpected system error"}
		render.Status(r, e.Code)
		render.JSON(w, r, e.Json())
	}
}

func renderApiError(w http.ResponseWriter, r *http.Request, err error, code int, msg string) {
	renderError(w, r, &Error{Err: err, Code: code, Msg: msg})
}
