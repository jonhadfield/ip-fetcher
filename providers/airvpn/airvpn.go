// Package airvpn retrieves AirVPN's published server ingress addresses.
package airvpn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "airvpn"
	FullName  = "AirVPN"
	HostType  = "anonymiser"
	SourceURL = "https://airvpn.org/"
	// DownloadURL returns AirVPN's status document, including each server's
	// public IPv4 and IPv6 ingress addresses.
	DownloadURL = "https://airvpn.org/api/status/"
)

var addressKeys = []string{
	"ip_v4_in1", "ip_v4_in2", "ip_v4_in3", "ip_v4_in4",
	"ip_v6_in1", "ip_v6_in2", "ip_v6_in3", "ip_v6_in4",
}

type AirVPN struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() AirVPN {
	return AirVPN{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawDoc struct {
	Servers []map[string]any `json:"servers"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (a *AirVPN) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, a.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download airvpn status from %s. http status code: %d", a.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (a *AirVPN) Fetch() (Doc, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw RawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	seen := make(map[netip.Prefix]struct{})
	var doc Doc

	for _, server := range raw.Servers {
		for _, key := range addressKeys {
			entry, _ := server[key].(string)
			if entry == "" {
				continue
			}

			prefix, ok := iplist.ToPrefix(entry)
			if !ok {
				continue
			}

			if _, exists := seen[prefix]; exists {
				continue
			}

			seen[prefix] = struct{}{}

			if prefix.Addr().Is4() {
				doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

				continue
			}

			doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
		}
	}

	slices.SortFunc(doc.IPv4Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })
	slices.SortFunc(doc.IPv6Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })

	return doc, nil
}
