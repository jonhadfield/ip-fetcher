// Package quiccloud retrieves QUIC.cloud's published CDN edge addresses.
package quiccloud

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "quiccloud"
	FullName  = "QUIC.cloud"
	HostType  = "cdn"
	SourceURL = "https://www.quic.cloud/ips"
	// DownloadURL is QUIC.cloud's list of CDN edge addresses. The page is a
	// single HTML fragment with addresses separated by <br /> tags rather than
	// a plain text file.
	DownloadURL = SourceURL
)

// errNoAddresses is returned when the page carries no parseable addresses.
var errNoAddresses = errors.New("failed to find any addresses on the quic.cloud ips page")

// brSplit matches the <br /> separators QUIC.cloud uses between addresses.
var brSplit = regexp.MustCompile(`(?i)<br\s*/?>`)

type QuicCloud struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() QuicCloud {
	return QuicCloud{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindAddresses returns the addresses embedded in the page, in order, with
// duplicates dropped.
func FindAddresses(page []byte) ([]string, error) {
	parts := brSplit.Split(string(page), -1)

	var addresses []string

	seen := make(map[string]struct{})

	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			continue
		}

		if _, ok := iplist.ToPrefix(entry); !ok {
			continue
		}

		if _, ok := seen[entry]; ok {
			continue
		}

		seen[entry] = struct{}{}

		addresses = append(addresses, entry)
	}

	if len(addresses) == 0 {
		return nil, errNoAddresses
	}

	return addresses, nil
}

// FetchData returns a newline separated list of addresses. The upstream page
// is HTML markup that is not worth publishing as-is.
func (q *QuicCloud) FetchData() ([]byte, http.Header, int, error) {
	if q.DownloadURL == "" {
		q.DownloadURL = DownloadURL
	}

	page, headers, status, err := web.Request(q.Client, q.DownloadURL, http.MethodGet, nil, nil, q.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download quic.cloud prefixes from %s. http status code: %d", q.DownloadURL, status)
	}

	addresses, err := FindAddresses(page)
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (q *QuicCloud) Fetch() (Doc, error) {
	data, _, _, err := q.FetchData()
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
