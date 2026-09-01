package youtube

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

type TrackInfo struct {
	ID        string
	Title     string `json:"title"`
	Channel   string `json:"author_name"`
	Thumbnail string `json:"thumbnail_url"`
}

func NewClient(apiKey string) *Client {
	cc := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Client{
		client:  cc,
		baseURL: "https://www.youtube.com",
		apiKey:  apiKey,
	}
}

func (c *Client) ExtractTrackInfo(videoID string) (*TrackInfo, error) {
	watchURL, err := url.Parse(c.baseURL + "/watch")
	if err != nil {
		return nil, ErrInvalidURL
	}

	watchQuery := watchURL.Query()
	watchQuery.Add("v", videoID)
	watchURL.RawQuery = watchQuery.Encode()

	oembedURL, err := url.JoinPath(c.baseURL, "oembed")
	if err != nil {
		return nil, ErrInvalidURL
	}

	reqURL, err := url.Parse(oembedURL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	q := reqURL.Query()
	q.Add("url", watchURL.String())
	q.Add("format", "json")
	reqURL.RawQuery = q.Encode()

	resp, err := c.client.Get(reqURL.String())
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	var info TrackInfo
	if err := dec.Decode(&info); err != nil {
		return nil, ErrRequestFailed
	}

	return &info, nil
}

const playlistItemsURL = "https://www.googleapis.com/youtube/v3/playlistItems"

// playlistItemsPageSize, Data API'nin izin verdiği maksimum sayfa boyutu.
const playlistItemsPageSize = 50

// maxPlaylistItems, tek importta çekilecek üst sınır — çok büyük
// playlist'lerde quota'yı ve tur/aday seçimini etkileyecek playlist
// şişkinliğini sınırlar.
const maxPlaylistItems = 500

// autoPlaylistPrefix, YouTube'un otomatik oluşturduğu karışık/radyo
// listelerinin ("Karışık liste" / mix, "Radio") playlist ID önekidir. Bu
// listeler sabit bir içerik değil, istek anında o an izlenen videoya ve
// isteği yapan kullanıcıya göre dinamik üretilir — aynı ID iki farklı
// istekte iki farklı sonuç dönebilir. Bu yüzden hiç import edilmiyor (bkz.
// FetchPlaylistItems).
const autoPlaylistPrefix = "RD"

type playlistItemsResponse struct {
	Items []struct {
		Snippet struct {
			Title                  string `json:"title"`
			ChannelTitle           string `json:"channelTitle"`
			VideoOwnerChannelTitle string `json:"videoOwnerChannelTitle"`
			ResourceId             struct {
				VideoId string `json:"videoId"`
			} `json:"resourceId"`
		} `json:"snippet"`
	} `json:"items"`
	NextPageToken string `json:"nextPageToken"`
}

// FetchPlaylistItems, bir YouTube playlist'indeki şarkıları (video id +
// başlık + kanal) döner. Silinmiş/gizli videolar (başlık veya videoId
// boş gelir) atlanır. Data API key (config.YoutubeAPIKey) gerektirir.
//
// YouTube'un otomatik karışık/radyo listeleri ("RD..", bkz. autoPlaylistPrefix)
// desteklenmez, ErrUnsupportedPlaylist döner. Diğer playlist'ler nextPageToken
// ile sonuna kadar (maxPlaylistItems'a kadar) sayfalanır.
func (c *Client) FetchPlaylistItems(playlistID string) ([]TrackInfo, error) {
	if strings.HasPrefix(playlistID, autoPlaylistPrefix) {
		return nil, ErrUnsupportedPlaylist
	}

	if c.apiKey == "" {
		return nil, ErrAPIKeyMissing
	}

	var tracks []TrackInfo
	pageToken := ""

	for {
		reqURL, err := url.Parse(playlistItemsURL)
		if err != nil {
			return nil, ErrInvalidURL
		}

		q := reqURL.Query()
		q.Add("part", "snippet")
		q.Add("playlistId", playlistID)
		q.Add("maxResults", strconv.Itoa(playlistItemsPageSize))
		q.Add("key", c.apiKey)
		if pageToken != "" {
			q.Add("pageToken", pageToken)
		}
		reqURL.RawQuery = q.Encode()

		resp, err := c.client.Get(reqURL.String())
		if err != nil {
			return nil, ErrRequestFailed
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, ErrPlaylistNotFound
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, ErrRequestFailed
		}

		var page playlistItemsResponse
		err = json.NewDecoder(resp.Body).Decode(&page)
		defer resp.Body.Close()
		if err != nil {
			return nil, ErrRequestFailed
		}

		for _, item := range page.Items {
			videoID := item.Snippet.ResourceId.VideoId
			if videoID == "" || item.Snippet.Title == "" {
				continue
			}

			channel := item.Snippet.VideoOwnerChannelTitle
			if channel == "" {
				channel = item.Snippet.ChannelTitle
			}

			tracks = append(tracks, TrackInfo{
				ID:      videoID,
				Title:   item.Snippet.Title,
				Channel: channel,
			})
		}

		if page.NextPageToken == "" || len(tracks) >= maxPlaylistItems {
			break
		}
		pageToken = page.NextPageToken
	}

	if len(tracks) > maxPlaylistItems {
		tracks = tracks[:maxPlaylistItems]
	}

	if len(tracks) == 0 {
		return nil, ErrPlaylistEmpty
	}

	return tracks, nil
}
