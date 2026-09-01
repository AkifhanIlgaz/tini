package playlist

import "github.com/AkifhanIlgaz/tini/internal/shared/htmx"

// AddLinkRequest is Handler.Add's input — a single YouTube link submitted
// through the top-of-page form.
type AddLinkRequest struct {
	URL string `form:"url"`
}

func (r AddLinkRequest) Validate() error {
	errs := htmx.FieldErrors{}

	if r.URL == "" {
		errs["url"] = ErrURLRequired
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
