// Package venue is the venue feature: the "Mekan bilgileri" page (QR kod +
// ayarlar formu) over a Mongo-backed repository. There's no way to create a
// venue through the app yet — cmd/seed writes the initial document directly
// — so Repository only supports reading and updating one that already
// exists.
package venue

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CollectionName is the Mongo collection Venue documents live in.
const CollectionName = "venues"

type VenueSettings struct {
	RoundIntervalMin int `bson:"round_interval_min" json:"roundIntervalMin"`
	CandidateCount   int `bson:"candidate_count" json:"candidateCount"`
	// RecentlyPlayedCooldownMin, bir şarkının tekrar aday/fallback
	// olabilmesi için son çalınışından bu kadar dakika geçmesi gerektiğini
	// belirtir (bkz. decisions.md 2026-07-26 — sayı bazlı pencereden süre
	// bazlı cooldown'a geçiş).
	RecentlyPlayedCooldownMin int `bson:"recently_played_cooldown_min" json:"recentlyPlayedCooldownMin"`
	// CandidateCooldownMin, bir şarkının round'da tekrar ADAY olabilmesi
	// için son aday olduğu andan bu kadar dakika geçmesi gerektiğini
	// belirtir. RecentlyPlayedCooldownMin'den ayrı: bir şarkı aday
	// gösterilip oy alamadan turu kaybedebilir (hiç çalınmamış olur) ama
	// yine de art arda turlarda sürekli aday olarak karşımıza çıkmasın
	// isteriz (bkz. decisions.md 2026-07-27).
	CandidateCooldownMin int `bson:"candidate_cooldown_min" json:"candidateCooldownMin"`
}

type Venue struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug       string        `bson:"slug" json:"slug"`
	Name       string        `bson:"name" json:"name"`
	LogoURL    string        `bson:"logo_url" json:"logoUrl"`
	Settings   VenueSettings `bson:"settings" json:"settings"`
	NowPlaying string        `bson:"now_playing" json:"nowPlaying"` // Youtube video ID of the currently playing track
	CreatedAt  time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt  time.Time     `bson:"updated_at" json:"updatedAt"`
}
