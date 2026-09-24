// Package surfshark retrieves Surfshark VPN egress addresses by resolving
// the hostnames published in their cluster catalogue.
package surfshark

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "surfshark"
	FullName  = "Surfshark"
	HostType  = "anonymiser"
	SourceURL = "https://surfshark.com/"
	// DownloadURL returns Surfshark's cluster catalogue. Each cluster only
	// publishes a hostname, so FetchData resolves those names to addresses.
	DownloadURL = "https://api.surfshark.com/v3/server/clusters/all"
)

// LookupIP resolves a hostname. Tests replace it so they need no network.
type LookupIP func(ctx context.Context, host string) ([]netip.Addr, error)

type Surfshark struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
	LookupIP    LookupIP
}

func New() Surfshark {
	return Surfshark{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
		LookupIP:    defaultLookupIP,
	}
}

func defaultLookupIP(ctx context.Context, host string) ([]netip.Addr, error) {
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	addrs := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip.IP)
		if !ok {
			continue
		}

		addrs = append(addrs, addr.Unmap())
	}

	return addrs, nil
}

type RawCluster struct {
	ConnectionName string `json:"connectionName"`
}

// ResolvedDoc is what FetchData publishes: hostnames plus the addresses they
// resolved to, so ProcessData can re-parse without the network.
type ResolvedDoc struct {
	Clusters []ResolvedCluster `json:"clusters"`
}

type ResolvedCluster struct {
	ConnectionName string   `json:"connection_name"`
	IPv4           []string `json:"ipv4"`
	IPv6           []string `json:"ipv6"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (s *Surfshark) FetchData() ([]byte, http.Header, int, error) {
	if s.DownloadURL == "" {
		s.DownloadURL = DownloadURL
	}

	if s.LookupIP == nil {
		s.LookupIP = defaultLookupIP
	}

	data, headers, status, err := web.Request(s.Client, s.DownloadURL, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download surfshark clusters from %s. http status code: %d", s.DownloadURL, status)
	}

	var clusters []RawCluster
	if err = json.Unmarshal(data, &clusters); err != nil {
		return nil, headers, status, err
	}

	ctx := context.Background()
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	resolved := ResolvedDoc{Clusters: make([]ResolvedCluster, 0, len(clusters))}
	for _, cluster := range clusters {
		if cluster.ConnectionName == "" {
			continue
		}

		entry := ResolvedCluster{ConnectionName: cluster.ConnectionName}

		addrs, lookupErr := s.LookupIP(ctx, cluster.ConnectionName)
		if lookupErr != nil {
			logrus.Warnf("failed to resolve surfshark host %s: %v", cluster.ConnectionName, lookupErr)
			resolved.Clusters = append(resolved.Clusters, entry)

			continue
		}

		for _, addr := range addrs {
			if addr.Is4() {
				entry.IPv4 = append(entry.IPv4, addr.String())

				continue
			}

			entry.IPv6 = append(entry.IPv6, addr.String())
		}

		resolved.Clusters = append(resolved.Clusters, entry)
	}

	out, err := json.MarshalIndent(resolved, "", "  ")
	if err != nil {
		return nil, headers, status, err
	}

	return append(out, '\n'), headers, status, nil
}

func (s *Surfshark) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var resolved ResolvedDoc
	if err := json.Unmarshal(data, &resolved); err != nil {
		return Doc{}, err
	}

	seen := make(map[netip.Prefix]struct{})
	var doc Doc

	add := func(entry string) {
		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
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

	for _, cluster := range resolved.Clusters {
		for _, addr := range cluster.IPv4 {
			add(addr)
		}

		for _, addr := range cluster.IPv6 {
			add(addr)
		}
	}

	slices.SortFunc(doc.IPv4Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })
	slices.SortFunc(doc.IPv6Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })

	return doc, nil
}
