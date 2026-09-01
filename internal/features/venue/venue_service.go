package venue

import (
	"context"
	"fmt"
)

type VenueService struct {
	repo VenueRepository
}

func NewService(repo VenueRepository) *VenueService {
	return &VenueService{repo: repo}
}

// GetVenue returns the venue identified by req.VenueID. req is assumed
// already validated.
func (s *VenueService) GetVenue(ctx context.Context, req GetVenueRequest) (Venue, error) {
	v, err := s.repo.FindByID(ctx, req.VenueID)
	if err != nil {
		return Venue{}, fmt.Errorf("venue: get venue: %w", err)
	}

	return v, nil
}

// UpdateVenue saves req's name and round settings onto req.VenueID. req is
// assumed already validated.
func (s *VenueService) UpdateVenue(ctx context.Context, req UpdateVenueRequest) (Venue, error) {
	v, err := s.repo.Update(ctx, req.VenueID, req.Name, req.ToVenueSettings())
	if err != nil {
		return Venue{}, fmt.Errorf("venue: update venue: %w", err)
	}

	return v, nil
}
