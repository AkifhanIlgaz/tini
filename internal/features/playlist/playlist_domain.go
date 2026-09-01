package playlist

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PlaylistItemsCollectionName is the Mongo collection PlaylistItem documents live in.
const PlaylistItemsCollectionName = "playlist_items"

// PlaylistItem is a single YouTube link added to a venue's playlist.
type PlaylistItem struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	YoutubeID string        `bson:"youtube_id" json:"youtubeId"`
	Title     string        `bson:"title" json:"title"`
	Channel   string        `bson:"channel" json:"channel"`
	VenueID   bson.ObjectID `bson:"venue_id" json:"-"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
	AddedBy   bson.ObjectID `bson:"added_by" json:"addedBy"`
	// LastPlayedAt, "son çalınanlar" cooldown filtresi için (bkz.
	// decisions.md 2026-07-26). Hiç çalınmadıysa nil.
	LastPlayedAt *time.Time `bson:"last_played_at" json:"lastPlayedAt,omitempty"`
	// LastCandidateAt, bir şarkının en son round adayı gösterildiği an
	// (kazanıp çalınmasa bile) — CandidateCooldownMin filtresi buna bakar
	// (bkz. decisions.md 2026-07-27). Hiç aday olmadıysa nil.
	LastCandidateAt *time.Time `bson:"last_candidate_at" json:"lastCandidateAt,omitempty"`
}
