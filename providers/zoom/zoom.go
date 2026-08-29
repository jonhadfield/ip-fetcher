package zoom

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
	ShortName = "zoom"
	FullName  = "Zoom"
	HostType  = "saas"
	SourceURL = "https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0060548"
	// DownloadURL returns the ranges Zoom's meeting and phone services connect
	// from, as a newline separated list of IPv4 prefixes.
	DownloadURL = "https://assets.zoom.us/docs/ipranges/Zoom.txt"
)

type Zoom struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Zoom {
	return Zoom{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (z *Zoom) FetchData() ([]byte, http.Header, int, error) {
	if z.DownloadURL == "" {
		z.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(z.Client, z.DownloadURL, http.MethodGet, nil, nil, z.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download zoom list from %s. http status code: %d", z.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (z *Zoom) Fetch() (Doc, error) {
	data, _, _, err := z.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData parses the list. Zoom publishes IPv4 only today, but the
// document is parsed by family so an IPv6 addition needs no change here.
func ProcessData(data []byte) (Doc, error) {
	ipv4, ipv6, err := iplist.Parse(ShortName, data)
	if err != nil {
		return Doc{}, err
	}

	return Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}, nil
}
