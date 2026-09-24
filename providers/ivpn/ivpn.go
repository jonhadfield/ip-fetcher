// Package ivpn retrieves IVPN's published VPN server ingress addresses.
package ivpn

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
	ShortName = "ivpn"
	FullName  = "IVPN"
	HostType  = "anonymiser"
	SourceURL = "https://www.ivpn.net/"
	// DownloadURL returns IVPN's WireGuard and OpenVPN server catalogue.
	DownloadURL = "https://api.ivpn.net/v5/servers.json"
)

type IVPN struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() IVPN {
	return IVPN{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawDoc struct {
	WireGuard []RawGateway `json:"wireguard"`
	OpenVPN   []RawGateway `json:"openvpn"`
}

type RawGateway struct {
	Hosts []RawHost `json:"hosts"`
}

type RawHost struct {
	Host  string `json:"host"`
	V2Ray string `json:"v2ray"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (i *IVPN) FetchData() ([]byte, http.Header, int, error) {
	if i.DownloadURL == "" {
		i.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(i.Client, i.DownloadURL, http.MethodGet, nil, nil, i.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download ivpn servers from %s. http status code: %d", i.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (i *IVPN) Fetch() (Doc, error) {
	data, _, _, err := i.FetchData()
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

	add := func(entry string) {
		if entry == "" {
			return
		}

		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
			return
		}

		// skip RFC1918 / ULA addresses published as tunnel local IPs
		if !prefix.Addr().IsGlobalUnicast() || prefix.Addr().IsPrivate() {
			return
		}

		if _, exists := seen[prefix]; exists {
			return
		}

		seen[prefix] = struct{}{}

		if prefix.Addr().Is4() {
			doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

			return
		}

		doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
	}

	for _, group := range [][]RawGateway{raw.WireGuard, raw.OpenVPN} {
		for _, gw := range group {
			for _, host := range gw.Hosts {
				add(host.Host)
				add(host.V2Ray)
			}
		}
	}

	slices.SortFunc(doc.IPv4Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })
	slices.SortFunc(doc.IPv6Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })

	return doc, nil
}
