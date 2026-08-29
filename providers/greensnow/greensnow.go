package greensnow

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
	ShortName = "greensnow"
	FullName  = "GreenSnow"
	HostType  = "threat"
	SourceURL = "https://greensnow.co/"
	// DownloadURL returns the hosts GreenSnow has observed attacking its
	// sensors, as a newline separated list of bare IPv4 addresses.
	DownloadURL = "https://blocklist.greensnow.co/greensnow.txt"
)

type GreenSnow struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() GreenSnow {
	return GreenSnow{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (g *GreenSnow) FetchData() ([]byte, http.Header, int, error) {
	if g.DownloadURL == "" {
		g.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(g.Client, g.DownloadURL, http.MethodGet, nil, nil, g.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download greensnow list from %s. http status code: %d", g.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (g *GreenSnow) Fetch() (Doc, error) {
	data, _, _, err := g.FetchData()
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
