package pingdom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "pingdom"
	FullName  = "Pingdom"
	HostType  = "monitoring"
	SourceURL = "https://www.pingdom.com/"
	// IPv4URL and IPv6URL return the addresses Pingdom's probe servers check
	// from, as newline separated lists of bare addresses.
	IPv4URL = "https://my.pingdom.com/probes/ipv4"
	IPv6URL = "https://my.pingdom.com/probes/ipv6"
)

type Pingdom struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() Pingdom {
	return Pingdom{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

// RawDoc is the combined representation of the two upstream lists, so both
// address families are stored in a single re-processable file.
type RawDoc struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (p *Pingdom) fetchList(url string) ([]string, http.Header, int, error) {
	data, headers, status, err := web.Request(p.Client, url, http.MethodGet, nil, nil, p.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download pingdom probes from %s. http status code: %d", url, status)
	}

	// the lists are bare addresses, so parse them by family and render back to
	// strings for the combined document.
	ipv4, ipv6, err := iplist.Parse(ShortName, data)
	if err != nil {
		return nil, headers, status, err
	}

	entries := make([]string, 0, len(ipv4)+len(ipv6))
	for _, prefix := range append(ipv4, ipv6...) {
		entries = append(entries, prefix.String())
	}

	return entries, headers, status, nil
}

func (p *Pingdom) FetchData() ([]byte, http.Header, int, error) {
	if p.IPv4URL == "" {
		p.IPv4URL = IPv4URL
	}

	if p.IPv6URL == "" {
		p.IPv6URL = IPv6URL
	}

	v4, headers, status, err := p.fetchList(p.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, _, v6Status, err := p.fetchList(p.IPv6URL)
	if err != nil {
		return nil, headers, v6Status, err
	}

	combined, err := json.MarshalIndent(RawDoc{IPv4: v4, IPv6: v6}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (p *Pingdom) Fetch() (Doc, error) {
	data, _, _, err := p.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var rawDoc RawDoc
	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return Doc{}, err
	}

	return Doc{
		IPv4Prefixes: castPrefixes(rawDoc.IPv4),
		IPv6Prefixes: castPrefixes(rawDoc.IPv6),
	}, nil
}

// castPrefixes parses the entries, logging and skipping any that do not parse
// rather than discarding the whole list.
func castPrefixes(in []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(in))

	for _, entry := range in {
		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse pingdom address: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	if len(prefixes) == 0 {
		return nil
	}

	return prefixes
}
