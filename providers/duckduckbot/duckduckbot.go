package duckduckbot

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "duckduckbot"
	FullName    = "DuckDuckBot"
	HostType    = "crawlers"
	SourceURL   = "https://duckduckgo.com/duckduckbot.json"
	DownloadURL = "https://duckduckgo.com/duckduckbot.json"
)

// The document format is shared with the other crawler providers, so the
// parsing lives in internal/botprefix. These are aliases, not new types, so
// this package's API is unchanged.

type (
	Doc          = botprefix.Doc
	RawDoc       = botprefix.RawDoc
	IPv4Entry    = botprefix.IPv4Entry
	IPv6Entry    = botprefix.IPv6Entry
	RawIPv4Entry = botprefix.RawIPv4Entry
	RawIPv6Entry = botprefix.RawIPv6Entry
)

type Duckduckbot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Duckduckbot {
	return Duckduckbot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (db *Duckduckbot) FetchData() ([]byte, http.Header, int, error) {
	if db.DownloadURL == "" {
		db.DownloadURL = DownloadURL
	}

	return web.Request(db.Client, db.DownloadURL, http.MethodGet, nil, nil, db.Timeout)
}

func (db *Duckduckbot) Fetch() (Doc, error) {
	data, _, _, err := db.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the feed. creationTime may be absent, so it is optional.
func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{})
}
