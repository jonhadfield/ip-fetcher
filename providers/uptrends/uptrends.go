package uptrends

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
	ShortName = "uptrends"
	FullName  = "Uptrends"
	HostType  = "monitoring"
	SourceURL = "https://www.uptrends.com/support/kb/account/ip-addresses-for-whitelisting"
	// IPv4URL and IPv6URL each return a json object whose Data member holds the
	// checkpoint addresses for that family, including ones announced but not yet
	// in service. Neither endpoint requires authentication.
	IPv4URL = "https://api.uptrends.com/v4/Checkpoint/Server/Ipv4"
	IPv6URL = "https://api.uptrends.com/v4/Checkpoint/Server/Ipv6"
)

type Uptrends struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() Uptrends {
	return Uptrends{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

// checkpointList is the upstream response for a single address family.
type checkpointList struct {
	Data []string `json:"Data"`
}

// RawDoc is the combined representation of the two upstream responses, so both
// families are stored in a single re-processable file.
type RawDoc struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (u *Uptrends) fetchList(url string) ([]string, http.Header, int, error) {
	headers := http.Header{"Accept": []string{"application/json"}}

	data, respHeaders, status, err := web.Request(u.Client, url, http.MethodGet, headers, nil, u.Timeout)
	if err != nil {
		return nil, respHeaders, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, respHeaders, status,
			fmt.Errorf("failed to download uptrends addresses from %s. http status code: %d", url, status)
	}

	var list checkpointList
	if err = json.Unmarshal(data, &list); err != nil {
		return nil, respHeaders, status, err
	}

	return list.Data, respHeaders, status, nil
}

func (u *Uptrends) FetchData() ([]byte, http.Header, int, error) {
	if u.IPv4URL == "" {
		u.IPv4URL = IPv4URL
	}

	if u.IPv6URL == "" {
		u.IPv6URL = IPv6URL
	}

	v4, headers, status, err := u.fetchList(u.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, _, v6Status, err := u.fetchList(u.IPv6URL)
	if err != nil {
		return nil, headers, v6Status, err
	}

	combined, err := json.MarshalIndent(RawDoc{IPv4: v4, IPv6: v6}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (u *Uptrends) Fetch() (Doc, error) {
	data, _, _, err := u.FetchData()
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

// castPrefixes turns the bare addresses the checkpoint endpoints return into
// host prefixes, tolerating a prefix should one ever appear.
func castPrefixes(in []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(in))

	for _, entry := range in {
		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse uptrends address: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	if len(prefixes) == 0 {
		return nil
	}

	return prefixes
}
