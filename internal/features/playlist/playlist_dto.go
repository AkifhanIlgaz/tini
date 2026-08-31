package playlist

import "errors"

// AddLinkRequest is Handler.Add's input — a single YouTube link submitted
// through the top-of-page form.
type AddLinkRequest struct {
	URL string `form:"url"`
}

func (r AddLinkRequest) Validate() error {
	if r.URL == "" {
		return errors.New("playlist: url is required")
	}

	return nil
}
