package perplexitybot

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "perplexitybot"
	FullName    = "PerplexityBot"
	HostType    = "crawlers"
	SourceURL   = "https://www.perplexity.com/perplexitybot.json"
	DownloadURL = "https://www.perplexity.com/perplexitybot.json"
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

type Perplexitybot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Perplexitybot {
	return Perplexitybot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (pb *Perplexitybot) FetchData() ([]byte, http.Header, int, error) {
	if pb.DownloadURL == "" {
		pb.DownloadURL = DownloadURL
	}

	return web.Request(pb.Client, pb.DownloadURL, http.MethodGet, nil, nil, pb.Timeout)
}

func (pb *Perplexitybot) Fetch() (Doc, error) {
	data, _, _, err := pb.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the feed. creationTime may be absent, so it is optional.
func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{})
}
