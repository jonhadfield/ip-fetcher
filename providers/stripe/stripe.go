package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"sort"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "stripe"
	FullName  = "Stripe"
	HostType  = "saas"
	SourceURL = "https://docs.stripe.com/ips"
	// WebhooksURL lists the IPs Stripe sends webhook notifications from.
	WebhooksURL = "https://stripe.com/files/ips/ips_webhooks.json"
	// APIURL lists the IPs Stripe's API is served from.
	APIURL = "https://stripe.com/files/ips/ips_api.json"
)

type Stripe struct {
	Client      *retryablehttp.Client
	WebhooksURL string
	APIURL      string
	Timeout     time.Duration
}

func New() Stripe {
	return Stripe{
		WebhooksURL: WebhooksURL,
		APIURL:      APIURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// RawDoc is the combined on-disk representation of the two upstream lists.
type RawDoc struct {
	Webhooks []string `json:"webhooks"`
	API      []string `json:"api"`
}

type Doc struct {
	Webhooks     []netip.Prefix `json:"webhooks"      yaml:"webhooks"`
	API          []netip.Prefix `json:"api"           yaml:"api"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (s *Stripe) fetchList(url string) ([]string, http.Header, int, error) {
	data, headers, status, err := web.Request(s.Client, url, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status, fmt.Errorf("failed to download stripe prefixes from %s. http status code: %d", url, status)
	}

	return collectStrings(data), headers, status, nil
}

// collectStrings extracts the IP strings whether the file is a flat array or an
// object whose single value is the array (e.g. {"WEBHOOKS": [...]}).
func collectStrings(data []byte) []string {
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		return list
	}

	var wrapped map[string][]string
	if err := json.Unmarshal(data, &wrapped); err == nil {
		var entries []string
		for _, values := range wrapped {
			entries = append(entries, values...)
		}

		return entries
	}

	return nil
}

func (s *Stripe) FetchData() ([]byte, http.Header, int, error) {
	if s.WebhooksURL == "" {
		s.WebhooksURL = WebhooksURL
	}

	if s.APIURL == "" {
		s.APIURL = APIURL
	}

	webhooks, headers, status, err := s.fetchList(s.WebhooksURL)
	if err != nil {
		return nil, headers, status, err
	}

	api, _, apiStatus, err := s.fetchList(s.APIURL)
	if err != nil {
		return nil, headers, apiStatus, err
	}

	combined, err := json.MarshalIndent(RawDoc{Webhooks: webhooks, API: api}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (s *Stripe) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
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
	doc.Webhooks = castEntries(rawDoc.Webhooks)
	doc.API = castEntries(rawDoc.API)

	seen4 := map[string]bool{}
	seen6 := map[string]bool{}

	for _, p := range append(append([]netip.Prefix{}, doc.Webhooks...), doc.API...) {
		if p.Addr().Is4() {
			if !seen4[p.String()] {
				seen4[p.String()] = true
				doc.IPv4Prefixes = append(doc.IPv4Prefixes, p)
			}

			continue
		}

		if !seen6[p.String()] {
			seen6[p.String()] = true
			doc.IPv6Prefixes = append(doc.IPv6Prefixes, p)
		}
	}

	sort.Slice(doc.IPv4Prefixes, func(i, j int) bool {
		return doc.IPv4Prefixes[i].String() < doc.IPv4Prefixes[j].String()
	})
	sort.Slice(doc.IPv6Prefixes, func(i, j int) bool {
		return doc.IPv6Prefixes[i].String() < doc.IPv6Prefixes[j].String()
	})

	return doc, nil
}

// castEntries parses Stripe entries, which are bare IP addresses or CIDRs.
func castEntries(entries []string) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, entry := range entries {
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			prefixes = append(prefixes, prefix)

			continue
		}

		if addr, err := netip.ParseAddr(entry); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))

			continue
		}

		logrus.Warnf("failed to parse stripe prefix: %s", entry)
	}

	return prefixes
}
