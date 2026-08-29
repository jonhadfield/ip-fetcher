package gcore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "gcore"
	FullName  = "Gcore CDN"
	HostType  = "cdn"
	SourceURL = "https://gcore.com/"
	// DownloadURL returns the edge addresses as json, with the two address
	// families in separate arrays.
	DownloadURL = "https://api.gcore.com/cdn/public-ip-list"
)

type Gcore struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Gcore {
	return Gcore{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawDoc struct {
	Addresses   []string `json:"addresses"`
	AddressesV6 []string `json:"addresses_v6"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (g *Gcore) FetchData() ([]byte, http.Header, int, error) {
	if g.DownloadURL == "" {
		g.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(g.Client, g.DownloadURL, http.MethodGet, nil, nil, g.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download gcore list from %s. http status code: %d", g.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (g *Gcore) Fetch() (Doc, error) {
	data, _, _, err := g.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var rawDoc RawDoc
	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return Doc{}, err
	}

	return Doc{
		IPv4Prefixes: castPrefixes(rawDoc.Addresses),
		IPv6Prefixes: castPrefixes(rawDoc.AddressesV6),
	}, nil
}

// castPrefixes parses the entries, logging and skipping any that do not parse
// rather than discarding the whole list.
func castPrefixes(in []string) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, entry := range in {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			logrus.Warnf("failed to parse gcore prefix: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes
}
