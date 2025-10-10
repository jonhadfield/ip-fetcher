package fastly

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	ShortName   = "fastly"
	FullName    = "Fastly"
	HostType    = "cdn"
	SourceURL   = "https://www.fastly.com/documentation/reference/api/utils/public-ip-list/"
	DownloadURL = "https://api.fastly.com/public-ip-list"
)

func New() Fastly {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Fastly{
		DownloadURL: DownloadURL,
		Client:      c,
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Fastly struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func (f *Fastly) FetchData() ([]byte, http.Header, int, error) {
	if f.DownloadURL == "" {
		f.DownloadURL = DownloadURL
	}

	return web.Request(f.Client, f.DownloadURL, http.MethodGet, nil, nil, f.Timeout)
}

func (f *Fastly) Fetch() (Doc, error) {
	data, _, _, err := f.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

type RawDoc struct {
	IPv4Addresses []string `json:"addresses"`
	IPv6Addresses []string `json:"ipv6_addresses"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"addresses"      yaml:"addresses"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_addresses" yaml:"ipv6_addresses"`
}

func castEntries(rd RawDoc) ([]netip.Prefix, []netip.Prefix) {
	var ipv4 []netip.Prefix
	var ipv6 []netip.Prefix

	for _, addr := range rd.IPv4Addresses {
		ipv4Prefix, parseError := netip.ParsePrefix(addr)
		if parseError == nil {
			ipv4 = append(ipv4, ipv4Prefix)

			continue
		}

		logrus.Warnf("failed to parse ipv4 prefix: %s", addr)
	}

	for _, addr := range rd.IPv6Addresses {
		ipv6Prefix, parseError := netip.ParsePrefix(addr)
		if parseError == nil {
			ipv6 = append(ipv6, ipv6Prefix)

			continue
		}

		logrus.Warnf("failed to parse ipv6 prefix: %s", addr)
	}

	return ipv4, ipv6
}

func ProcessData(data []byte) (Doc, error) {
	var rawDoc RawDoc
	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return Doc{}, err
	}

	doc := Doc{}
	doc.IPv4Prefixes, doc.IPv6Prefixes = castEntries(rawDoc)

	return doc, nil
}
