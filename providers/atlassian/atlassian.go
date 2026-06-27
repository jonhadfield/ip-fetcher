package atlassian

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
	ShortName   = "atlassian"
	FullName    = "Atlassian"
	HostType    = "saas"
	SourceURL   = "https://ip-ranges.atlassian.com/"
	DownloadURL = "https://ip-ranges.atlassian.com/"
)

type Atlassian struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Atlassian {
	return Atlassian{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// Item mirrors an entry in the upstream "items" array.
type Item struct {
	Network   string   `json:"network"   yaml:"network"`
	MaskLen   int      `json:"mask_len"  yaml:"mask_len"`
	CIDR      string   `json:"cidr"      yaml:"cidr"`
	Mask      string   `json:"mask"      yaml:"mask"`
	Region    []string `json:"region"    yaml:"region"`
	Product   []string `json:"product"   yaml:"product"`
	Direction []string `json:"direction" yaml:"direction"`
}

type Doc struct {
	CreationDate string         `json:"creationDate"  yaml:"creationDate"`
	SyncToken    string         `json:"syncToken"     yaml:"syncToken"`
	Items        []Item         `json:"items"         yaml:"items"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (a *Atlassian) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, a.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status, fmt.Errorf("failed to download atlassian prefixes. http status code: %d", status)
	}

	return data, headers, status, nil
}

func (a *Atlassian) Fetch() (Doc, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

type rawDoc struct {
	CreationDate string `json:"creationDate"`
	SyncToken    string `json:"syncToken"`
	Items        []Item `json:"items"`
}

func ProcessData(data []byte) (Doc, error) {
	var raw rawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{
		CreationDate: raw.CreationDate,
		SyncToken:    raw.SyncToken,
		Items:        raw.Items,
	}

	for _, item := range raw.Items {
		p, perr := netip.ParsePrefix(item.CIDR)
		if perr != nil {
			logrus.Warnf("failed to parse atlassian prefix: %s", item.CIDR)

			continue
		}

		if p.Addr().Is4() {
			doc.IPv4Prefixes = append(doc.IPv4Prefixes, p)

			continue
		}

		doc.IPv6Prefixes = append(doc.IPv6Prefixes, p)
	}

	return doc, nil
}
