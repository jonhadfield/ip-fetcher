package ccbot

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/botprefix"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "ccbot"
	FullName    = "Common Crawl CCBot"
	HostType    = "crawlers"
	SourceURL   = "https://commoncrawl.org/ccbot"
	DownloadURL = "https://index.commoncrawl.org/ccbot.json"
)

type (
	Doc          = botprefix.Doc
	RawDoc       = botprefix.RawDoc
	IPv4Entry    = botprefix.IPv4Entry
	IPv6Entry    = botprefix.IPv6Entry
	RawIPv4Entry = botprefix.RawIPv4Entry
	RawIPv6Entry = botprefix.RawIPv6Entry
)

type CCBot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() CCBot {
	return CCBot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (c *CCBot) FetchData() ([]byte, http.Header, int, error) {
	if c.DownloadURL == "" {
		c.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(c.Client, c.DownloadURL, http.MethodGet, nil, nil, c.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download ccbot prefixes from %s. http status code: %d", c.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (c *CCBot) Fetch() (Doc, error) {
	data, _, _, err := c.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return botprefix.Parse(data, botprefix.Options{
		TimeFormats:         []string{time.RFC3339, botprefix.DefaultTimeFormat},
		RequireCreationTime: true,
	})
}
