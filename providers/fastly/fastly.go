package fastly

import (
	"encoding/json"
	"net/http"
	"net/netip"

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
	}
}

type Fastly struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func (f *Fastly) FetchData() (data []byte, headers http.Header, status int, err error) {
	if f.DownloadURL == "" {
		f.DownloadURL = DownloadURL
	}

	return web.Request(f.Client, f.DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (f *Fastly) Fetch() (doc Doc, err error) {
	data, _, _, err := f.FetchData()
	if err != nil {
		return
	}

	return ProcessData(data)
}

type RawDoc struct {
	IPv4Addresses []string `json:"addresses"`
	IPv6Addresses []string `json:"ipv6_addresses"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"addresses" yaml:"addresses"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_addresses" yaml:"ipv6_addresses"`
}

func castEntries(rd RawDoc) (ipv4 []netip.Prefix, ipv6 []netip.Prefix) {
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

func ProcessData(data []byte) (doc Doc, err error) {
	var rawDoc RawDoc
	err = json.Unmarshal(data, &rawDoc)
	if err != nil {
		return
	}

	doc.IPv4Prefixes, doc.IPv6Prefixes = castEntries(rawDoc)
	if err != nil {
		return
	}

	return
}
