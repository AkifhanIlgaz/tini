package youtube

import "net/url"

const youtubeHost = "www.youtube.com"
const youtubeShortHost = "youtu.be"
const youtubeWatchPath = "/watch"
const youtubePlaylistPath = "/playlist"
const youtubeVideoIDParam = "v"
const youtubePlaylistIDParam = "list"

// ParsedURL, bir YouTube linkinden çıkarılabilecek video ve/veya playlist
// kimlikleridir. İkisi de dolu olabilir (ör. bir playlist içindeki videonun
// linki, `?v=...&list=...`) — bu durumda hangisinin ekleneceğine çağıran
// karar verir (bkz. track.AddTrackRequest.Mode).
type ParsedURL struct {
	VideoID    string
	PlaylistID string
}

// ParseURL, desteklenen YouTube link biçimlerinden (izleme linki, kısa link,
// playlist linki) video/playlist kimliklerini çıkarır. Hiçbiri bulunamazsa
// ErrInvalidURL döner. Playlist tipi (kullanıcı playlist'i "PL" ya da
// YouTube'un otomatik karışık/radyo listesi "RD") burada ayrıştırılmaz —
// FetchPlaylistItems importta buna göre farklı davranır (bkz. client.go).
func ParseURL(youtubeURL string) (ParsedURL, error) {
	u, err := url.Parse(youtubeURL)
	if err != nil {
		return ParsedURL{}, ErrInvalidURL
	}

	var parsed ParsedURL

	switch u.Host {
	case youtubeShortHost:
		parsed.VideoID = trimLeadingSlash(u.Path)
	case youtubeHost:
		switch u.Path {
		case youtubeWatchPath:
			parsed.VideoID = u.Query().Get(youtubeVideoIDParam)
			parsed.PlaylistID = u.Query().Get(youtubePlaylistIDParam)
		case youtubePlaylistPath:
			parsed.PlaylistID = u.Query().Get(youtubePlaylistIDParam)
		}
	}

	if parsed.VideoID == "" && parsed.PlaylistID == "" {
		return ParsedURL{}, ErrInvalidURL
	}

	return parsed, nil
}

func trimLeadingSlash(path string) string {
	if len(path) > 0 && path[0] == '/' {
		return path[1:]
	}
	return path
}
