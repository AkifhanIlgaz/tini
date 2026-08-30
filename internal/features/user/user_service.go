package user

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// InviteAdmin pre-provisions an admin account from req — the admin's own
// name/avatar are filled in on their first Google login (see
// Repository.Upsert). req is assumed already validated (see the handler's
// req.Validate() call).
func (s *Service) InviteAdmin(ctx context.Context, req InviteAdminRequest) (User, error) {
	created, err := s.repo.CreateAdmin(ctx, req.ToUser())
	if err != nil {
		return User{}, fmt.Errorf("user: invite admin: %w", err)
	}

	return created, nil
}

// RemoveAdmin deletes the admin identified by req, but only if it's
// actually an admin of req.VenueID — a boss can't remove an admin
// belonging to another venue. req is assumed already validated.
func (s *Service) RemoveAdmin(ctx context.Context, req RemoveAdminRequest) error {
	if err := s.repo.DeleteAdmin(ctx, req.VenueID, req.AdminID); err != nil {
		return fmt.Errorf("user: remove admin: %w", err)
	}

	return nil
}

// ListUsers returns every user (boss and admins alike) belonging to
// req.VenueID. req is assumed already validated.
func (s *Service) ListUsers(ctx context.Context, req ListUsersRequest) ([]User, error) {
	users, err := s.repo.ListUsersByVenue(ctx, req.VenueID)
	if err != nil {
		return nil, fmt.Errorf("user: list users: %w", err)
	}

	return users, nil
}
