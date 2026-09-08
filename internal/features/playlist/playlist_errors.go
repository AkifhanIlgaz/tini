package playlist

import "errors"

var (
	ErrItemAlreadyExists = errors.New("playlist: item already exists")
	ErrItemNotFound      = errors.New("playlist: item not found")
)

// AddLinkRequest.Validate's field errors, plus the display counterpart
// Handler.Add uses for the ErrItemAlreadyExists conflict above — their
// Error() text is shown to the user directly via field.FieldError (see
// htmx.FieldErrors), unlike the sentinel errors above.
var (
	ErrURLRequired     = errors.New("Link boş olamaz.")
	ErrURLAlreadyAdded = errors.New("Bu şarkı zaten playlist'te.")
)
