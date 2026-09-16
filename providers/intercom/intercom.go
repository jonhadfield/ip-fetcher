package intercom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"sort"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "intercom"
	FullName  = "Intercom"
	HostType  = "saas"
	SourceURL = "https://developers.intercom.com/docs/build-an-integration/learn-more/ip-allowlisting"
	USURL     = "https://static.intercomcdn.com/intercom-ips/us/intercom-ip-ranges.json"
	EUURL     = "https://static.intercomcdn.com/intercom-ips/eu/intercom-ip-ranges.json"
	AUURL     = "https://static.intercomcdn.com/intercom-ips/au/intercom-ip-ranges.json"
)

type Intercom struct {
	Client  *retryablehttp.Client
	USURL   string
	EUURL   string
	AUURL   string
	Timeout time.Duration
}

func New() Intercom {
	return Intercom{
		USURL:   USURL,
		EUURL:   EUURL,
		AUURL:   AUURL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

type RawRange struct {
	Range   string `json:"range"`
	Region  string `json:"region"`
	Service string `json:"service"`
}

type RawDoc struct {
	IPRanges []RawRange `json:"ip_ranges"`
}

type Doc struct {
	IPRanges     []RawRange     `json:"ip_ranges"     yaml:"ip_ranges"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (i *Intercom) fetchRegion(downloadURL string) (RawDoc, http.Header, int, error) {
	data, headers, status, err := web.Request(i.Client, downloadURL, http.MethodGet, nil, nil, i.Timeout)
	if err != nil {
		return RawDoc{}, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return RawDoc{}, headers, status,
			fmt.Errorf("failed to download intercom ranges from %s. http status code: %d", downloadURL, status)
	}

	var raw RawDoc
	if err = json.Unmarshal(data, &raw); err != nil {
		return RawDoc{}, headers, status, err
	}

	return raw, headers, status, nil
}

// FetchData returns the three regional documents merged into one. Intercom
// publishes each workspace region separately.
func (i *Intercom) FetchData() ([]byte, http.Header, int, error) {
	if i.USURL == "" {
		i.USURL = USURL
	}

	if i.EUURL == "" {
		i.EUURL = EUURL
	}

	if i.AUURL == "" {
		i.AUURL = AUURL
	}

	us, headers, status, err := i.fetchRegion(i.USURL)
	if err != nil {
		return nil, headers, status, err
	}

	eu, _, euStatus, err := i.fetchRegion(i.EUURL)
	if err != nil {
		return nil, headers, euStatus, err
	}

	au, _, auStatus, err := i.fetchRegion(i.AUURL)
	if err != nil {
		return nil, headers, auStatus, err
	}

	combined := RawDoc{IPRanges: append(append(us.IPRanges, eu.IPRanges...), au.IPRanges...)}

	data, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return nil, headers, status, err
	}

	return data, headers, status, nil
}

func (i *Intercom) Fetch() (Doc, error) {
	data, _, _, err := i.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw RawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{IPRanges: raw.IPRanges}
	seen4 := map[string]struct{}{}
	seen6 := map[string]struct{}{}

	for _, entry := range raw.IPRanges {
		prefix, ok := iplist.ToPrefix(entry.Range)
		if !ok {
			continue
		}

		if prefix.Addr().Is4() {
			if _, exists := seen4[prefix.String()]; exists {
				continue
			}

			seen4[prefix.String()] = struct{}{}
			doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

			continue
		}

		if _, exists := seen6[prefix.String()]; exists {
			continue
		}

		seen6[prefix.String()] = struct{}{}
		doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
	}

	sort.Slice(doc.IPv4Prefixes, func(i, j int) bool {
		return doc.IPv4Prefixes[i].String() < doc.IPv4Prefixes[j].String()
	})
	sort.Slice(doc.IPv6Prefixes, func(i, j int) bool {
		return doc.IPv6Prefixes[i].String() < doc.IPv6Prefixes[j].String()
	})

	return doc, nil
}
