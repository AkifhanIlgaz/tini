package htmx

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// ToastError is a handler's way of describing an expected, user-facing
// failure (bad input, a conflicting record, ...) without calling Toast
// itself — HandleError is the single place that turns it into one, so every
// handler reports the same way instead of each building its own toast.
type ToastError struct {
	Title       string
	Description string
}

func (e *ToastError) Error() string {
	return e.Title
}

// NewToastError builds a ToastError — return it directly from a handler
// instead of calling Toast and returning nil.
func NewToastError(title, description string) error {
	return &ToastError{Title: title, Description: description}
}

// FieldErrors is a DTO's Validate() failure mode for bad form input — one
// error per invalid field, keyed by its form name, instead of a single
// ToastError covering every mistake. A feature's own <feature>_errors.go
// defines the actual field errors (their Error() text is shown to the user
// directly via field.FieldError, unlike the app's other sentinel errors,
// which are internal and never rendered as-is). A handler builds a
// FieldErrors with errors.As(err, &fieldErrs) and re-renders the form
// instead of calling Toast.
type FieldErrors map[string]error

func (e FieldErrors) Error() string {
	return "htmx: invalid fields"
}

// HandleError is the app's single fiber.ErrorHandler (wired in
// cmd/server/main.go) — the one place that decides how an error returned
// from a handler reaches the client, so handlers themselves never call
// Toast for an error case. A ToastError is an expected, already-described
// failure, so it just renders as its own toast.
//
// Any other error is unexpected, and is logged here before being converted
// to a response — not "left to slog-fiber" as one might expect: slog-fiber
// derives its own log line from *this function's return value*, not from
// the err it was called with, so once HandleError "handles" an error (by
// turning it into a 200 toast, or letting fiber.DefaultErrorHandler turn it
// into a plain-text response) the original wrapped error and its %w chain
// are gone by the time slog-fiber would otherwise log them. This is the
// same reasoning as cmd/server/main.go's slog.Error before os.Exit: there's
// no higher layer left for the error to be wrapped-and-returned to, so it's
// logged here instead (see CONTRIBUTING.md's single handling rule).
func HandleError(c fiber.Ctx, err error) error {
	var toastErr *ToastError
	if errors.As(err, &toastErr) {
		return Toast(c, ToastOptions{
			Title:       toastErr.Title,
			Description: toastErr.Description,
			Variant:     ToastDanger,
		})
	}

	slog.Error("unhandled request error", "error", err, "method", c.Method(), "path", c.Path())

	if IsRequest(c) {
		if err := Toast(c, ToastOptions{
			Title:       "Bir şeyler ters gitti",
			Description: "Lütfen daha sonra tekrar deneyin.",
			Variant:     ToastDanger,
		}); err != nil {
			return err
		}
	}

	return fiber.DefaultErrorHandler(c, err)
}
