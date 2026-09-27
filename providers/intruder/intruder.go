// Package intruder retrieves Intruder vulnerability scanner addresses.
package intruder

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
	ShortName = "intruder"
	FullName  = "Intruder"
	HostType  = "scanner"
	SourceURL = "https://help.intruder.io/en/articles/1635683-what-ips-do-i-need-to-add-to-my-allowlist"
	// DownloadURL is the help article listing the scanner ranges. There is no
	// machine readable feed, so the addresses are taken from the page.
	DownloadURL = SourceURL
)

// errNoAddresses is returned when the page carries no addresses, which means it
// has been restructured.
var errNoAddresses = errors.New("failed to find any addresses on the intruder allowlist page")

// addressRegexp matches the marked up cells the ranges are published in.
var addressRegexp = regexp.MustCompile(`(?s)<code[^>]*>([^<]+)</code>`)

type Intruder struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Intruder {
	return Intruder{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindAddresses returns the addresses the page publishes, in the order they
// appear, with any repeats dropped.
func FindAddresses(page []byte) ([]string, error) {
	var addresses []string

	seen := make(map[string]struct{})

	for _, match := range addressRegexp.FindAllSubmatch(page, -1) {
		entry := strings.TrimSpace(string(match[1]))

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
// carries markup that changes with every help-centre build, so it is not worth
// publishing.
func (i *Intruder) FetchData() ([]byte, http.Header, int, error) {
	if i.DownloadURL == "" {
		i.DownloadURL = DownloadURL
	}

	page, headers, status, err := web.Request(i.Client, i.DownloadURL, http.MethodGet, nil, nil, i.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download intruder addresses from %s. http status code: %d", i.DownloadURL, status)
	}

	addresses, err := FindAddresses(page)
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (i *Intruder) Fetch() (Doc, error) {
	data, _, _, err := i.FetchData()
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
