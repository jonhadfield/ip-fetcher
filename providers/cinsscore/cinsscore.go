package cinsscore

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
	ShortName = "cinsscore"
	FullName  = "CINS Army List"
	HostType  = "threat"
	SourceURL = "https://cinsscore.com/"
	// DownloadURL returns the CI Army bad guys list as a newline separated list
	// of bare IPv4 addresses.
	DownloadURL = "https://cinsscore.com/list/ci-badguys.txt"
)

type CINSScore struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() CINSScore {
	return CINSScore{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (c *CINSScore) FetchData() ([]byte, http.Header, int, error) {
	if c.DownloadURL == "" {
		c.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(c.Client, c.DownloadURL, http.MethodGet, nil, nil, c.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download cinsscore list from %s. http status code: %d", c.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (c *CINSScore) Fetch() (Doc, error) {
	data, _, _, err := c.FetchData()
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
