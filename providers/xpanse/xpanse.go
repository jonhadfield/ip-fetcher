// Package xpanse retrieves Palo Alto Cortex Xpanse internet-scanning ranges.
package xpanse

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
	ShortName = "xpanse"
	FullName  = "Cortex Xpanse"
	HostType  = "scanner"
	SourceURL = "https://cortex-docs.paloaltonetworks.com/cortex-xpanse/reference/scanning-activity"
	// DownloadURL is the documentation page listing the ranges used for Cortex
	// Xpanse scanning. There is no machine readable feed, so the addresses are
	// taken from the page.
	DownloadURL = SourceURL
)

// errNoAddresses is returned when the page carries no addresses, which means it
// has been restructured.
var errNoAddresses = errors.New("failed to find any addresses on the cortex xpanse scanning page")

// cidrRegexp matches IPv4 and IPv6 CIDRs. Bare addresses are ignored: the page
// publishes prefixes, and short colon-separated markup would otherwise match.
var cidrRegexp = regexp.MustCompile(
	`\b(?:(?:\d{1,3}\.){3}\d{1,3}/\d{1,2}|(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}/\d{1,3})\b`,
)

type Xpanse struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Xpanse {
	return Xpanse{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindAddresses returns the CIDRs the page publishes, in the order they appear,
// with any repeats dropped.
func FindAddresses(page []byte) ([]string, error) {
	var addresses []string

	seen := make(map[string]struct{})

	for _, match := range cidrRegexp.FindAllString(string(page), -1) {
		entry := strings.TrimSpace(match)

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

// FetchData returns the addresses as a newline separated list: the page itself
// carries markup that changes with every documentation build, so it is not
// worth publishing.
func (x *Xpanse) FetchData() ([]byte, http.Header, int, error) {
	if x.DownloadURL == "" {
		x.DownloadURL = DownloadURL
	}

	page, headers, status, err := web.Request(x.Client, x.DownloadURL, http.MethodGet, nil, nil, x.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download cortex xpanse addresses from %s. http status code: %d", x.DownloadURL, status)
	}

	addresses, err := FindAddresses(page)
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (x *Xpanse) Fetch() (Doc, error) {
	data, _, _, err := x.FetchData()
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
