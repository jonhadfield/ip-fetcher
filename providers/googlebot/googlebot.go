package googlebot

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "googlebot"
	FullName    = "Google Crawler Bots"
	HostType    = "crawlers"
	SourceURL   = "https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot"
	DownloadURL = "https://developers.google.com/static/crawling/ipranges/common-crawlers.json"
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

type Googlebot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Googlebot {
	return Googlebot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (gc *Googlebot) FetchData() ([]byte, http.Header, int, error) {
	if gc.DownloadURL == "" {
		gc.DownloadURL = DownloadURL
	}

	return web.Request(gc.Client, gc.DownloadURL, http.MethodGet, nil, nil, gc.Timeout)
}

func (gc *Googlebot) Fetch() (Doc, error) {
	data, _, _, err := gc.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the feed. Google always publishes creationTime, so its
// absence is an error.
func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{RequireCreationTime: true})
}
