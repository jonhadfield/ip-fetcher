package googleuta

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "googleuta"
	FullName    = "Google User-Triggered Agents"
	HostType    = "crawlers"
	SourceURL   = "https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot"
	DownloadURL = "https://developers.google.com/static/crawling/ipranges/user-triggered-agents.json"
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

type Googleuta struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Googleuta {
	return Googleuta{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (ga *Googleuta) FetchData() ([]byte, http.Header, int, error) {
	if ga.DownloadURL == "" {
		ga.DownloadURL = DownloadURL
	}

	return web.Request(ga.Client, ga.DownloadURL, http.MethodGet, nil, nil, ga.Timeout)
}

func (ga *Googleuta) Fetch() (Doc, error) {
	data, _, _, err := ga.FetchData()
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
