package anthropic

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "anthropic"
	FullName    = "Anthropic Crawler Bots"
	HostType    = "crawlers"
	SourceURL   = "https://claude.com/crawling/bots.json"
	DownloadURL = "https://claude.com/crawling/bots.json"
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

type Anthropic struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Anthropic {
	return Anthropic{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (a *Anthropic) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, a.Timeout)
}

func (a *Anthropic) Fetch() (Doc, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the feed. creationTime is published as RFC3339 here,
// while sibling feeds use a fractional seconds variant, so both are accepted
// and it is optional.
func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{TimeFormats: []string{time.RFC3339, botprefix.DefaultTimeFormat}})
}
