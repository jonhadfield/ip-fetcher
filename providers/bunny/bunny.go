package bunny

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "bunny"
	FullName  = "Bunny.net"
	HostType  = "cdn"
	SourceURL = "https://bunny.net/"
	// IPv4URL returns a JSON array of the edge server IPv4 addresses.
	IPv4URL = "https://api.bunny.net/system/edgeserverlist"
	// IPv6URL returns a JSON array of the edge server IPv6 addresses.
	IPv6URL = "https://api.bunny.net/system/edgeserverlist/ipv6"
)

type Bunny struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() Bunny {
	return Bunny{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

// RawDoc is the combined representation of the two upstream endpoints. It is the
// on-disk format used by the publisher so that both address families are stored
// in a single, re-fetchable file.
type RawDoc struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (b *Bunny) fetchList(url string) ([]string, http.Header, int, error) {
	data, headers, status, err := web.Request(b.Client, url, http.MethodGet, nil, nil, b.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status, fmt.Errorf("failed to download bunny prefixes from %s. http status code: %d", url, status)
	}

	var entries []string
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, headers, status, err
	}

	return entries, headers, status, nil
}

func (b *Bunny) FetchData() ([]byte, http.Header, int, error) {
	if b.IPv4URL == "" {
		b.IPv4URL = IPv4URL
	}

	if b.IPv6URL == "" {
		b.IPv6URL = IPv6URL
	}

	v4, headers, status, err := b.fetchList(b.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, _, v6status, err := b.fetchList(b.IPv6URL)
	if err != nil {
		return nil, headers, v6status, err
	}

	combined, err := json.MarshalIndent(RawDoc{IPv4: v4, IPv6: v6}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (b *Bunny) Fetch() (Doc, error) {
	data, _, _, err := b.FetchData()
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

	doc := Doc{}
	doc.IPv4Prefixes = castEntries(rawDoc.IPv4)
	doc.IPv6Prefixes = castEntries(rawDoc.IPv6)

	return doc, nil
}

// castEntries parses the edge server list, which may contain either bare IP
// addresses or CIDR prefixes, into netip.Prefix values.
func castEntries(entries []string) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, entry := range entries {
		prefix, ok := toPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse bunny prefix: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes
}

// toPrefix accepts either a CIDR prefix or a bare IP address and returns a
// netip.Prefix (host prefixes are used for bare addresses).
func toPrefix(entry string) (netip.Prefix, bool) {
	if prefix, err := netip.ParsePrefix(entry); err == nil {
		return prefix, true
	}

	if addr, err := netip.ParseAddr(entry); err == nil {
		return netip.PrefixFrom(addr, addr.BitLen()), true
	}

	return netip.Prefix{}, false
}
