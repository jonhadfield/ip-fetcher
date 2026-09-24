// Package x4bnet retrieves the X4BNet community list of commercial VPN prefixes.
package x4bnet

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "x4bnet"
	FullName  = "X4BNet VPN"
	HostType  = "anonymiser"
	SourceURL = "https://github.com/X4BNet/lists_vpn"
	// IPv4URL is the VPN-only IPv4 prefix list (not the broader datacenter list).
	IPv4URL = "https://raw.githubusercontent.com/X4BNet/lists_vpn/main/output/vpn/ipv4.txt"
	// IPv6URL is the matching IPv6 prefix list.
	IPv6URL = "https://raw.githubusercontent.com/X4BNet/lists_vpn/main/output/vpn/ipv6.txt"
)

type X4BNet struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() X4BNet {
	return X4BNet{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (x *X4BNet) fetchList(url string) ([]byte, http.Header, int, error) {
	data, headers, status, err := web.Request(x.Client, url, http.MethodGet, nil, nil, x.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download x4bnet list from %s. http status code: %d", url, status)
	}

	return data, headers, status, nil
}

// FetchData returns a combined document of both address families so a saved
// copy can be re-parsed without a second request.
func (x *X4BNet) FetchData() ([]byte, http.Header, int, error) {
	if x.IPv4URL == "" {
		x.IPv4URL = IPv4URL
	}

	if x.IPv6URL == "" {
		x.IPv6URL = IPv6URL
	}

	v4, headers, status, err := x.fetchList(x.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, _, status6, err := x.fetchList(x.IPv6URL)
	if err != nil {
		return nil, headers, status6, err
	}

	data, err := iplist.MarshalFamilies(lines(v4), lines(v6))
	if err != nil {
		return nil, headers, status, err
	}

	return data, headers, status, nil
}

func (x *X4BNet) Fetch() (Doc, error) {
	data, _, _, err := x.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	ipv4, ipv6, err := iplist.ParseFamilies(ShortName, data)
	if err != nil {
		return Doc{}, err
	}

	return Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}, nil
}

func lines(data []byte) []string {
	var out []string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		entry := strings.TrimSpace(scanner.Text())
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}

		out = append(out, entry)
	}

	return out
}
