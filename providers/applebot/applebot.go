package applebot

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "applebot"
	FullName    = "Applebot"
	HostType    = "crawlers"
	SourceURL   = "https://support.apple.com/en-us/119829"
	DownloadURL = "https://search.developer.apple.com/applebot.json"
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

type Applebot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Applebot {
	return Applebot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (ab *Applebot) FetchData() ([]byte, http.Header, int, error) {
	if ab.DownloadURL == "" {
		ab.DownloadURL = DownloadURL
	}

	return web.Request(ab.Client, ab.DownloadURL, http.MethodGet, nil, nil, ab.Timeout)
}

func (ab *Applebot) Fetch() (Doc, error) {
	data, _, _, err := ab.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the feed. creationTime may be absent, so it is optional.
func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{})
}
