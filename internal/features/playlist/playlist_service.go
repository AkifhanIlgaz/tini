package playlist

import (
	"context"
	"fmt"
)

type PlaylistService struct {
	repo PlaylistRepository
}

func NewService(repo PlaylistRepository) *PlaylistService {
	return &PlaylistService{repo: repo}
}

// ListItems returns req.VenueID's playlist items. req is assumed already
// validated.
func (s *PlaylistService) ListItems(ctx context.Context, req ListItemsRequest) ([]PlaylistItem, error) {
	items, err := s.repo.FindByVenueID(ctx, req.VenueID)
	if err != nil {
		return nil, fmt.Errorf("playlist: list items: %w", err)
	}

	return items, nil
}

// AddItem saves req as a new playlist item on req.VenueID. req is assumed
// already validated.
func (s *PlaylistService) AddItem(ctx context.Context, req AddLinkRequest) (PlaylistItem, error) {
	item, err := s.repo.Insert(ctx, PlaylistItem{
		VenueID: req.VenueID,
		VideoID: req.VideoID,
		URL:     req.URL,
	})
	if err != nil {
		return PlaylistItem{}, fmt.Errorf("playlist: add item: %w", err)
	}

	return item, nil
}

// DeleteItem removes req.ID from req.VenueID's playlist. req is assumed
// already validated.
func (s *PlaylistService) DeleteItem(ctx context.Context, req DeleteItemRequest) error {
	if err := s.repo.Delete(ctx, req.VenueID, req.ID); err != nil {
		return fmt.Errorf("playlist: delete item: %w", err)
	}

	return nil
}
