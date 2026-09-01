// Command seed creates one demo venue plus its boss and two dummy admins —
// a one-off dev/ops tool, not something the running server ever calls.
// The venue feature (repository/service/handler) isn't built yet, so this
// writes the venue document directly rather than through an abstraction
// that doesn't exist; user documents go through the real user.UserRepository
// since that one already fits.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/AkifhanIlgaz/tini/internal/config"
	"github.com/AkifhanIlgaz/tini/internal/features/user"
	"github.com/AkifhanIlgaz/tini/internal/features/venue"
	db "github.com/AkifhanIlgaz/tini/internal/platform/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// bossEmail already went through Google login during manual testing — it
// exists as a boss with no venue yet, so this seed attaches the venue to
// it rather than creating it fresh (falling back to creation only if it's
// somehow missing).
const bossEmail = "akifhanilgazz@gmail.com"

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config.Load", "error", err)
		os.Exit(1)
	}

	mongoClient, err := db.Connect(ctx, cfg.Mongo)
	if err != nil {
		slog.Error("mongo.Connect", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			slog.Error("mongo.Disconnect", "error", err)
		}
	}()

	venueID, err := createVenue(ctx, mongoClient)
	if err != nil {
		slog.Error("create venue", "error", err)
		os.Exit(1)
	}
	slog.Info("venue created", "id", venueID.Hex())

	usersRepo, err := user.NewRepository(mongoClient)
	if err != nil {
		slog.Error("user.NewRepository", "error", err)
		os.Exit(1)
	}

	if err := attachBoss(ctx, usersRepo, venueID); err != nil {
		slog.Error("attach boss", "error", err)
		os.Exit(1)
	}

	admins := []struct {
		email string
		name  string
	}{
		{email: "admin1@tini.test", name: "Admin Bir"},
		{email: "admin2@tini.test", name: "Admin İki"},
	}

	for _, admin := range admins {
		created, err := usersRepo.CreateAdmin(ctx, user.User{
			Email:   admin.email,
			Name:    admin.name,
			Role:    user.RoleAdmin,
			VenueID: venueID,
		})
		if err != nil {
			slog.Error("create admin", "email", admin.email, "error", err)
			os.Exit(1)
		}
		slog.Info("admin created", "id", created.ID.Hex(), "email", created.Email)
	}
}

func createVenue(ctx context.Context, client *db.Client) (bson.ObjectID, error) {
	now := time.Now()

	v := venue.Venue{
		Slug: "demo",
		Name: "Demo Mekan",
		Settings: venue.VenueSettings{
			RoundIntervalMin:          5,
			CandidateCount:            4,
			RecentlyPlayedCooldownMin: 60,
			CandidateCooldownMin:      30,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := client.Database().Collection(venue.VenueCollectionName).InsertOne(ctx, v)
	if err != nil {
		return bson.ObjectID{}, err
	}

	return result.InsertedID.(bson.ObjectID), nil
}

// attachBoss sets venueID on the account identified by bossEmail, creating
// it as a boss first if it doesn't exist yet.
func attachBoss(ctx context.Context, repo user.UserRepository, venueID bson.ObjectID) error {
	existing, err := repo.FindByEmail(ctx, bossEmail)
	if errors.Is(err, user.ErrUserNotFound) {
		// CreateAdmin just inserts whatever User it's given — despite the
		// name, nothing in it forces Role to admin, so it works fine here.
		created, err := repo.CreateAdmin(ctx, user.User{
			Email:   bossEmail,
			Role:    user.RoleBoss,
			VenueID: venueID,
		})
		if err != nil {
			return err
		}

		slog.Info("boss created", "id", created.ID.Hex(), "email", created.Email)

		return nil
	}
	if err != nil {
		return err
	}

	if err := repo.SetVenueID(ctx, existing.ID, venueID); err != nil {
		return err
	}

	slog.Info("boss attached to venue", "id", existing.ID.Hex(), "email", existing.Email)

	return nil
}
