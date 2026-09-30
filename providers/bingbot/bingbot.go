package bingbot

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "bingbot"
	FullName    = "Bingbot"
	HostType    = "crawlers"
	SourceURL   = "https://www.bing.com/webmasters/help/how-to-verify-bingbot-3905dc26"
	DownloadURL = "https://www.bing.com/toolbox/bingbot.json"
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

type Bingbot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Bingbot {
	return Bingbot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (bb *Bingbot) FetchData() ([]byte, http.Header, int, error) {
	if bb.DownloadURL == "" {
		bb.DownloadURL = DownloadURL
	}

	return web.Request(bb.Client, bb.DownloadURL, http.MethodGet, nil, nil, bb.Timeout)
}

func (bb *Bingbot) Fetch() (Doc, error) {
	data, _, _, err := bb.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the feed. Bing always publishes creationTime, so its
// absence is an error.
func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{RequireCreationTime: true})
}
