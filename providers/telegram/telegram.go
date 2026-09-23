// Package telegram retrieves Telegram's published DC and Bot API address ranges.
package telegram

import (
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "telegram"
	FullName  = "Telegram"
	HostType  = "saas"
	SourceURL = "https://core.telegram.org/resources/cidr.txt"
	// DownloadURL is Telegram's newline separated list of DC and Bot API
	// prefixes, used when allowlisting webhook delivery.
	DownloadURL = SourceURL
)

type Telegram struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Telegram {
	return Telegram{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (t *Telegram) FetchData() ([]byte, http.Header, int, error) {
	if t.DownloadURL == "" {
		t.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(t.Client, t.DownloadURL, http.MethodGet, nil, nil, t.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download telegram prefixes from %s. http status code: %d", t.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (t *Telegram) Fetch() (Doc, error) {
	data, _, _, err := t.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	ipv4, ipv6, err := iplist.Parse(ShortName, data)
	if err != nil {
		return Doc{}, err
	}

	return Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}, nil
}
