// Package invicti retrieves Invicti (Acunetix) cloud scanner addresses.
package invicti

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
	ShortName = "invicti"
	FullName  = "Invicti"
	HostType  = "scanner"
	SourceURL = "https://docs.invicti.com/ip/trustlist-us"
	// USDownloadURL and EUDownloadURL are the regional trustlist pages. There is
	// no machine readable feed, so the addresses are taken from both pages.
	USDownloadURL = SourceURL
	EUDownloadURL = "https://docs.invicti.com/ip/trustlist-eu"
)

// errNoAddresses is returned when the pages carry no addresses, which means
// they have been restructured.
var errNoAddresses = errors.New("failed to find any addresses on the invicti trustlist pages")

// cellRegexp matches table cells that hold the published sources.
var cellRegexp = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)

// tagRegexp strips markup inside a cell so only the text remains.
var tagRegexp = regexp.MustCompile(`<[^>]+>`)

// addressRegexp finds bare addresses or CIDRs, including those wrapped in
// parentheses next to a scanner hostname.
var addressRegexp = regexp.MustCompile(
	`\b(?:(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?|(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}(?:/\d{1,3})?)\b`,
)

type Invicti struct {
	Client        *retryablehttp.Client
	USDownloadURL string
	EUDownloadURL string
	Timeout       time.Duration
}

func New() Invicti {
	return Invicti{
		USDownloadURL: USDownloadURL,
		EUDownloadURL: EUDownloadURL,
		Client:        web.NewHTTPClientWithLogger(),
		Timeout:       web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindAddresses returns the addresses published in table cells, in the order
// they appear, with any repeats dropped.
func FindAddresses(page []byte) ([]string, error) {
	var addresses []string

	seen := make(map[string]struct{})

	for _, match := range cellRegexp.FindAllSubmatch(page, -1) {
		plain := tagRegexp.ReplaceAllString(string(match[1]), " ")

		for _, entry := range addressRegexp.FindAllString(plain, -1) {
			if _, ok := iplist.ToPrefix(entry); !ok {
				continue
			}

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

func (i *Invicti) fetchPage(url string) ([]byte, http.Header, int, error) {
	page, headers, status, err := web.Request(i.Client, url, http.MethodGet, nil, nil, i.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download invicti addresses from %s. http status code: %d", url, status)
	}

	return page, headers, status, nil
}

// FetchData returns the addresses as a newline separated list merged from the
// US and EU trustlist pages.
func (i *Invicti) FetchData() ([]byte, http.Header, int, error) {
	if i.USDownloadURL == "" {
		i.USDownloadURL = USDownloadURL
	}

	if i.EUDownloadURL == "" {
		i.EUDownloadURL = EUDownloadURL
	}

	usPage, headers, status, err := i.fetchPage(i.USDownloadURL)
	if err != nil {
		return nil, headers, status, err
	}

	euPage, _, euStatus, err := i.fetchPage(i.EUDownloadURL)
	if err != nil {
		return nil, headers, euStatus, err
	}

	addresses, err := FindAddresses(append(append([]byte{}, usPage...), euPage...))
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (i *Invicti) Fetch() (Doc, error) {
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
