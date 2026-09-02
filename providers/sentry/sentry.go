package sentry

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
	ShortName = "sentry"
	FullName  = "Sentry Uptime"
	HostType  = "monitoring"
	SourceURL = "https://docs.sentry.io/security-legal-pii/security/ip-ranges/"
	// DownloadURL returns the addresses Sentry's uptime checks run from, as a
	// newline separated list of bare addresses.
	DownloadURL = "https://sentry.io/api/0/uptime-ips/"
)

type Sentry struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Sentry {
	return Sentry{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (s *Sentry) FetchData() ([]byte, http.Header, int, error) {
	if s.DownloadURL == "" {
		s.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(s.Client, s.DownloadURL, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download sentry addresses from %s. http status code: %d", s.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (s *Sentry) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
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
