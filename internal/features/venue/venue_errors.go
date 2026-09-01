package venue

import "errors"

var ErrVenueNotFound = errors.New("venue: not found")

// UpdateVenueRequest.Validate's field errors — their Error() text is shown
// to the user directly via field.FieldError (see htmx.FieldErrors), unlike
// ErrVenueNotFound above.
var (
	ErrNameRequired                     = errors.New("Mekan adını boş bırakamazsın.")
	ErrRoundIntervalMinInvalid          = errors.New("Tur aralığı 0'dan büyük olmalı.")
	ErrCandidateCountInvalid            = errors.New("Aday sayısı 0'dan büyük olmalı.")
	ErrRecentlyPlayedCooldownMinInvalid = errors.New("Son çalınanlar bekleme süresi 0'dan büyük olmalı.")
	ErrCandidateCooldownMinInvalid      = errors.New("Aday bekleme süresi 0'dan büyük olmalı.")
)
