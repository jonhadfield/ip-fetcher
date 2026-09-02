package detectify

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
	ShortName = "detectify"
	FullName  = "Detectify"
	HostType  = "scanner"
	SourceURL = "https://docs.detectify.com/network-setup/scanner-ip-addresses"
	// DownloadURL is the documentation page listing the addresses Detectify's
	// surface monitoring, application scanning and connectors run from. There is
	// no machine readable feed, so the addresses are taken from the page.
	DownloadURL = SourceURL
)

// errNoAddresses is returned when the page carries no addresses, which means it
// has been restructured.
var errNoAddresses = errors.New("failed to find any addresses on the detectify scanner page")

// addressRegexp matches the marked up cells the addresses are published in. The
// page's prose and its icons hold digits that look like addresses, so only the
// cells are considered.
var addressRegexp = regexp.MustCompile(`(?s)<code[^>]*>([^<]+)</code>`)

type Detectify struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Detectify {
	return Detectify{
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
// appear, with any repeated between sections dropped.
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
// carries markup that changes with every documentation build, so it is not
// worth publishing.
func (d *Detectify) FetchData() ([]byte, http.Header, int, error) {
	if d.DownloadURL == "" {
		d.DownloadURL = DownloadURL
	}

	page, headers, status, err := web.Request(d.Client, d.DownloadURL, http.MethodGet, nil, nil, d.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download detectify addresses from %s. http status code: %d", d.DownloadURL, status)
	}

	addresses, err := FindAddresses(page)
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (d *Detectify) Fetch() (Doc, error) {
	data, _, _, err := d.FetchData()
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
