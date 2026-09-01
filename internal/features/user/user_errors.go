package user

import "errors"

var (
	ErrUserNotFound      = errors.New("user: not found")
	ErrUserAlreadyExists = errors.New("user: already exists")
)

// InviteAdminRequest.Validate's field errors, plus the display counterpart
// Handler.InviteAdmin uses for the ErrUserAlreadyExists conflict above —
// their Error() text is shown to the user directly via field.FieldError
// (see htmx.FieldErrors), unlike the sentinel errors above.
var (
	ErrEmailInvalid       = errors.New("Geçerli bir e-posta adresi gir.")
	ErrEmailAlreadyExists = errors.New("Bu e-posta zaten kayıtlı.")
)
