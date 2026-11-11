package googleutfg

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "googleutfg"
	FullName    = "Google User-Triggered Fetchers (Google)"
	HostType    = "crawlers"
	SourceURL   = "https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot"
	DownloadURL = "https://developers.google.com/static/crawling/ipranges/user-triggered-fetchers-google.json"
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

type Googleutfg struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Googleutfg {
	return Googleutfg{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (gf *Googleutfg) FetchData() ([]byte, http.Header, int, error) {
	if gf.DownloadURL == "" {
		gf.DownloadURL = DownloadURL
	}

	return web.Request(gf.Client, gf.DownloadURL, http.MethodGet, nil, nil, gf.Timeout)
}

func (gf *Googleutfg) Fetch() (Doc, error) {
	data, _, _, err := gf.FetchData()
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
