package imperva

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
	ShortName = "imperva"
	FullName  = "Imperva"
	HostType  = "security"
	SourceURL = "https://docs.imperva.com/bundle/cloud-application-security/page/more/restricting-direct-access.htm"
	// DownloadURL is Imperva's public IP list endpoint. It is a POST endpoint
	// that returns JSON when resp_format=json is supplied.
	DownloadURL = "https://my.incapsula.com/api/integration/v1/ips?resp_format=json"
)

type Imperva struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Imperva {
	return Imperva{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (i *Imperva) FetchData() ([]byte, http.Header, int, error) {
	if i.DownloadURL == "" {
		i.DownloadURL = DownloadURL
	}

	// The endpoint is a POST; resp_format=json is passed in the query string so
	// no request body is required.
	data, headers, status, err := web.Request(i.Client, i.DownloadURL, http.MethodPost, nil, nil, i.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status, fmt.Errorf("failed to download imperva prefixes. http status code: %d", status)
	}

	return data, headers, status, nil
}

func (i *Imperva) Fetch() (Doc, error) {
	data, _, _, err := i.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

type rawDoc struct {
	IPRanges   []string `json:"ipRanges"`
	IPv6Ranges []string `json:"ipv6Ranges"`
}

func ProcessData(data []byte) (Doc, error) {
	var raw rawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{}
	doc.IPv4Prefixes = castEntries(raw.IPRanges)
	doc.IPv6Prefixes = castEntries(raw.IPv6Ranges)

	return doc, nil
}

// castEntries parses Imperva entries, which are CIDRs or bare IP addresses.
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

		logrus.Warnf("failed to parse imperva prefix: %s", entry)
	}

	return prefixes
}
