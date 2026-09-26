// Package nodeping retrieves NodePing uptime monitoring probe addresses.
package nodeping

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "nodeping"
	FullName  = "NodePing"
	HostType  = "monitoring"
	SourceURL = "https://nodeping.com/FAQ"
	// DownloadURL returns hostname and address pairs for each probe. The same
	// addresses are also published via DNS on probes.nodeping.com.
	DownloadURL = "https://nodeping.com/content/txt/pinghosts.txt"
)

type NodePing struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() NodePing {
	return NodePing{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (n *NodePing) FetchData() ([]byte, http.Header, int, error) {
	if n.DownloadURL == "" {
		n.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(n.Client, n.DownloadURL, http.MethodGet, nil, nil, n.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download nodeping addresses from %s. http status code: %d", n.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (n *NodePing) Fetch() (Doc, error) {
	data, _, _, err := n.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	seen := make(map[netip.Prefix]struct{})
	var doc Doc

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		for _, field := range strings.Fields(line) {
			prefix, ok := iplist.ToPrefix(field)
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

	if err := scanner.Err(); err != nil {
		return Doc{}, err
	}

	slices.SortFunc(doc.IPv4Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })
	slices.SortFunc(doc.IPv6Prefixes, func(a, b netip.Prefix) int { return a.Addr().Compare(b.Addr()) })

	return doc, nil
}
