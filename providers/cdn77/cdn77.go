package cdn77

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
	ShortName   = "cdn77"
	FullName    = "CDN77"
	HostType    = "cdn"
	SourceURL   = "https://www.cdn77.com/"
	DownloadURL = "https://prefixlists.tools.cdn77.com/public_lmax_prefixes.json"
)

type CDN77 struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() CDN77 {
	return CDN77{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (c *CDN77) FetchData() ([]byte, http.Header, int, error) {
	if c.DownloadURL == "" {
		c.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(c.Client, c.DownloadURL, http.MethodGet, nil, nil, c.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status, fmt.Errorf("failed to download cdn77 prefixes. http status code: %d", status)
	}

	return data, headers, status, nil
}

func (c *CDN77) Fetch() (Doc, error) {
	data, _, _, err := c.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	entries, err := parseEntries(data)
	if err != nil {
		return Doc{}, err
	}

	doc := Doc{}

	for _, entry := range entries {
		prefix, ok := toPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse cdn77 prefix: %s", entry)

			continue
		}

		if prefix.Addr().Is4() {
			doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

			continue
		}

		doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
	}

	return doc, nil
}

// parseEntries accepts the prefix list whether it is a flat JSON array of
// strings or an object whose string array values hold the prefixes.
func parseEntries(data []byte) ([]string, error) {
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil
	}

	var wrapped map[string][]string
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}

	var entries []string
	for _, values := range wrapped {
		entries = append(entries, values...)
	}

	return entries, nil
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
