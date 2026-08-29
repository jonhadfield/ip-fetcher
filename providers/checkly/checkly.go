package checkly

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
	ShortName = "checkly"
	FullName  = "Checkly"
	HostType  = "monitoring"
	SourceURL = "https://www.checklyhq.com/docs/monitoring/allowlisting/"
	// IPv4URL returns bare addresses and IPv6URL returns prefixes, each as a
	// json array, so the two families come from separate endpoints.
	IPv4URL = "https://api.checklyhq.com/v1/static-ips"
	IPv6URL = "https://api.checklyhq.com/v1/static-ipv6s"
)

type Checkly struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() Checkly {
	return Checkly{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

// RawDoc is the combined representation of the two upstream arrays, so both
// families are stored in a single re-processable file.
type RawDoc struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (c *Checkly) fetchList(url string) ([]string, http.Header, int, error) {
	data, headers, status, err := web.Request(c.Client, url, http.MethodGet, nil, nil, c.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download checkly addresses from %s. http status code: %d", url, status)
	}

	var entries []string
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, headers, status, err
	}

	return entries, headers, status, nil
}

func (c *Checkly) FetchData() ([]byte, http.Header, int, error) {
	if c.IPv4URL == "" {
		c.IPv4URL = IPv4URL
	}

	if c.IPv6URL == "" {
		c.IPv6URL = IPv6URL
	}

	v4, headers, status, err := c.fetchList(c.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, _, v6Status, err := c.fetchList(c.IPv6URL)
	if err != nil {
		return nil, headers, v6Status, err
	}

	combined, err := json.MarshalIndent(RawDoc{IPv4: v4, IPv6: v6}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (c *Checkly) Fetch() (Doc, error) {
	data, _, _, err := c.FetchData()
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

// castPrefixes accepts both notations: the IPv4 endpoint returns bare
// addresses, which become host prefixes, while the IPv6 endpoint returns
// prefixes already.
func castPrefixes(in []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(in))

	for _, entry := range in {
		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse checkly address: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	if len(prefixes) == 0 {
		return nil
	}

	return prefixes
}
