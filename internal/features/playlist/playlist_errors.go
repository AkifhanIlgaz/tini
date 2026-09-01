package playlist

import "errors"

// AddLinkRequest.Validate's field error — its Error() text is shown to the
// user directly via field.FieldError (see htmx.FieldErrors).
var ErrURLRequired = errors.New("Link boş olamaz.")
