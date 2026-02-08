package akamai

import (
	"bufio"
	"bytes"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "akamai"
	FullName    = "Akamai"
	HostType    = "cdn"
	SourceURL   = "https://techdocs.akamai.com/"
	DownloadURL = "https://ip-ranges.akamai.com/"
)

type Akamai struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Akamai {
	return Akamai{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (a *Akamai) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, a.Timeout)
}

func (a *Akamai) Fetch() ([]netip.Prefix, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return nil, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) ([]netip.Prefix, error) {
	r := bytes.NewReader(data)
	scanner := bufio.NewScanner(r)
	var prefixes []netip.Prefix
	for scanner.Scan() {
		prefix, err := netip.ParsePrefix(scanner.Text())
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}
