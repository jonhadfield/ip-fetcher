// Package rapid7 retrieves Rapid7 InsightAppSec cloud engine scanner addresses.
package rapid7

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
	ShortName = "rapid7"
	FullName  = "Rapid7 InsightAppSec"
	HostType  = "scanner"
	SourceURL = "https://docs.rapid7.com/insightappsec/allowlist-cloud-engine-ips/"
	// DownloadURL is the documentation page listing the InsightAppSec cloud
	// engine addresses by region. There is no machine readable feed, so the
	// addresses are taken from the page.
	DownloadURL = SourceURL
)

// errNoAddresses is returned when the page carries no addresses, which means it
// has been restructured.
var errNoAddresses = errors.New("failed to find any addresses on the rapid7 insightappsec page")

// cellRegexp matches table cells. The page also embeds a large React payload
// whose version strings look like addresses, so only rendered cells are read.
var cellRegexp = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)

// tagRegexp strips markup inside a cell so only the text remains.
var tagRegexp = regexp.MustCompile(`<[^>]+>`)

type Rapid7 struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Rapid7 {
	return Rapid7{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindAddresses returns the addresses published in table cells whose contents
// are only addresses, in the order they appear, with repeats dropped.
func FindAddresses(page []byte) ([]string, error) {
	var addresses []string

	seen := make(map[string]struct{})

	for _, match := range cellRegexp.FindAllSubmatch(page, -1) {
		plain := strings.Fields(tagRegexp.ReplaceAllString(string(match[1]), " "))
		if len(plain) == 0 {
			continue
		}

		var cell []string

		for _, token := range plain {
			if _, ok := iplist.ToPrefix(token); !ok {
				cell = nil

				break
			}

			cell = append(cell, token)
		}

		for _, entry := range cell {
			if _, ok := seen[entry]; ok {
				continue
			}

			seen[entry] = struct{}{}

			addresses = append(addresses, entry)
		}
	}

	if len(addresses) == 0 {
		return nil, errNoAddresses
	}

	return addresses, nil
}

// FetchData returns the addresses as a newline separated list: the page itself
// carries markup that changes with every documentation build, so it is not
// worth publishing.
func (r *Rapid7) FetchData() ([]byte, http.Header, int, error) {
	if r.DownloadURL == "" {
		r.DownloadURL = DownloadURL
	}

	page, headers, status, err := web.Request(r.Client, r.DownloadURL, http.MethodGet, nil, nil, r.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download rapid7 addresses from %s. http status code: %d", r.DownloadURL, status)
	}

	addresses, err := FindAddresses(page)
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (r *Rapid7) Fetch() (Doc, error) {
	data, _, _, err := r.FetchData()
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
