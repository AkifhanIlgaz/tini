package youtube

import "errors"

var (
	ErrInvalidURL       = errors.New("Lütfen geçerli bir YouTube URL'si giriniz.")
	ErrRequestFailed    = errors.New("Video bilgilerini almaya çalışırken bir hata oluştu.")
	ErrAPIKeyMissing    = errors.New("Playlist eklemek için YouTube API anahtarı yapılandırılmamış.")
	ErrPlaylistNotFound = errors.New("Playlist bulunamadı veya herkese açık değil.")
	ErrPlaylistEmpty    = errors.New("Playlist'te eklenebilecek şarkı bulunamadı.")
	// ErrUnsupportedPlaylist, "RD" ile başlayan playlist'ler için döner.
	// Bunlar YouTube'un istek anında, o an izlenen videoya ve kullanıcıya göre
	// dinamik ürettiği karışık/radyo listeleridir — sabit bir içeriği yok,
	// API'den her çekildiğinde farklı sonuç dönebilir, bu yüzden import
	// edilmiyor (bkz. FetchPlaylistItems).
	ErrUnsupportedPlaylist = errors.New("Bu bir karışık/radyo listesi — YouTube bunu isteğe ve kullanıcıya göre anlık oluşturduğu için içeri aktarılamıyor. Sadece kullanıcı oynatma listeleri (\"PL\" ile başlayan) eklenebilir.")
)
