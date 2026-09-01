package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/AkifhanIlgaz/tini/internal/platform/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const usersCollectionName = "users"

type UserRepository interface {
	CreateAdmin(ctx context.Context, u User) (User, error)
	ListUsersByVenue(ctx context.Context, venueID bson.ObjectID) ([]User, error)
	DeleteAdmin(ctx context.Context, venueID, adminID bson.ObjectID) error
	FindByID(ctx context.Context, id bson.ObjectID) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Upsert(ctx context.Context, email, name, avatarURL string) (User, error)
	SetVenueID(ctx context.Context, userID, venueID bson.ObjectID) error
}

type userMongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(client *db.Client) (UserRepository, error) {
	collection := client.Database().Collection(usersCollectionName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "venue_id", Value: 1},
				{Key: "role", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"role": RoleBoss}),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return nil, fmt.Errorf("user: create indexes: %w", err)
	}

	return &userMongoRepository{
		collection: collection,
	}, nil
}

// CreateAdmin inserts u as a pre-provisioned admin — the caller is
// responsible for setting Email/VenueID/Role before an invited admin has
// ever logged in via Google.
func (r *userMongoRepository) CreateAdmin(ctx context.Context, u User) (User, error) {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return User{}, ErrUserAlreadyExists
		}

		return User{}, fmt.Errorf("user: create admin: %w", err)
	}

	u.ID = result.InsertedID.(bson.ObjectID)

	return u, nil
}

// DeleteAdmin deletes adminID, but only if it's actually an admin of
// venueID — the filter itself is the ownership check, so a boss can't
// remove an admin belonging to another venue.
func (r *userMongoRepository) DeleteAdmin(ctx context.Context, venueID, adminID bson.ObjectID) error {
	filter := bson.M{
		"_id":      adminID,
		"venue_id": venueID,
		"role":     RoleAdmin,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("user: delete admin: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsersByVenue returns every user (boss and admins alike) belonging to
// venueID.
func (r *userMongoRepository) ListUsersByVenue(ctx context.Context, venueID bson.ObjectID) ([]User, error) {
	filter := bson.M{
		"venue_id": venueID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("user: list users by venue: %w", err)
	}

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("user: list users by venue: %w", err)
	}

	return users, nil
}

// SetVenueID attaches venueID to the user identified by userID — used once
// a boss finishes creating their venue.
func (r *userMongoRepository) SetVenueID(ctx context.Context, userID, venueID bson.ObjectID) error {
	filter := bson.M{
		"_id": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"venue_id":   venueID,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("user: set venue id: %w", err)
	}

	return nil
}

func (r *userMongoRepository) FindByID(ctx context.Context, id bson.ObjectID) (User, error) {
	var u User

	filter := bson.M{
		"_id": id,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("user: find by id: %w", err)
	}

	return u, nil
}

func (r *userMongoRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User

	filter := bson.M{
		"email": email,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("user: find by email: %w", err)
	}

	return u, nil
}

// Upsert creates the user on first login, or updates name/avatar on
// subsequent logins — email is the stable key Google OAuth gives us.
// Roles is left untouched on update: role changes are managed separately,
// not overwritten by whatever Google returns.
func (r *userMongoRepository) Upsert(ctx context.Context, email, name, avatarURL string) (User, error) {
	var u User

	now := time.Now()

	filter := bson.M{
		"email": email,
	}

	update := bson.M{
		"$set": bson.M{
			"name":       name,
			"avatar_url": avatarURL,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"email":      email,
			"created_at": now,
			"role":       RoleBoss,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&u)
	if err != nil {
		return User{}, fmt.Errorf("user: upsert: %w", err)
	}

	return u, nil
}
