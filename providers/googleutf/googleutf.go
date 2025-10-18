package googleutf

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName                = "googleutf"
	FullName                 = "Google User-Triggered Fetchers"
	HostType                 = "crawlers"
	SourceURL                = "https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot"
	DownloadURL              = "https://developers.google.com/static/search/apis/ipranges/user-triggered-fetchers.json"
	downloadedFileTimeFormat = "2006-01-02T15:04:05.999999"
)

func New() Googleutf {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Googleutf{
		DownloadURL: DownloadURL,
		Client:      c,
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Googleutf struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

type RawDoc struct {
	CreationTime  string            `json:"creationTime"`
	LastRequested time.Time         `json:"-" yaml:"-"`
	Entries       []json.RawMessage `json:"prefixes"`
}

func (gu *Googleutf) FetchData() ([]byte, http.Header, int, error) {
	if gu.DownloadURL == "" {
		gu.DownloadURL = DownloadURL
	}
	return web.Request(gu.Client, gu.DownloadURL, http.MethodGet, nil, nil, gu.Timeout)
}

func (gu *Googleutf) Fetch() (Doc, error) {
	data, _, _, err := gu.FetchData()
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

	ipv4, ipv6, err := castEntries(rawDoc.Entries)
	if err != nil {
		return Doc{}, err
	}

	ct, err := time.Parse(downloadedFileTimeFormat, rawDoc.CreationTime)
	if err != nil {
		return Doc{}, err
	}

	return Doc{
		CreationTime: ct,
		IPv4Prefixes: ipv4,
		IPv6Prefixes: ipv6,
	}, nil
}

func castEntries(prefixes []json.RawMessage) ([]IPv4Entry, []IPv6Entry, error) {
	var ipv4 []IPv4Entry
	var ipv6 []IPv6Entry
	for _, pr := range prefixes {
		var ipv4entry RawIPv4Entry

		var ipv6entry RawIPv6Entry

		// try 4
		if err := json.Unmarshal(pr, &ipv4entry); err == nil {
			ipv4Prefix, parseError := netip.ParsePrefix(ipv4entry.IPv4Prefix)
			if parseError == nil {
				ipv4 = append(ipv4, IPv4Entry{
					IPv4Prefix: ipv4Prefix,
				})

				continue
			}
		}

		// try 6
		ipv6Err := json.Unmarshal(pr, &ipv6entry)
		if ipv6Err == nil {
			ipv6Prefix, parseError := netip.ParsePrefix(ipv6entry.IPv6Prefix)
			if parseError != nil {
				return ipv4, ipv6, parseError
			}

			ipv6 = append(ipv6, IPv6Entry{
				IPv6Prefix: ipv6Prefix,
			})

			continue
		}

		return ipv4, ipv6, ipv6Err
	}

	return ipv4, ipv6, nil
}

type RawIPv4Entry struct {
	IPv4Prefix string `json:"ipv4Prefix"`
}

type RawIPv6Entry struct {
	IPv6Prefix string `json:"ipv6Prefix"`
}

type IPv4Entry struct {
	IPv4Prefix netip.Prefix `json:"ipv4Prefix"`
}

type IPv6Entry struct {
	IPv6Prefix netip.Prefix `json:"ipv6Prefix"`
}

type Doc struct {
	CreationTime time.Time   `json:"creationTime" yaml:"creationTime"`
	IPv4Prefixes []IPv4Entry `json:"ipv4Prefixes" yaml:"ipv4Prefixes"`
	IPv6Prefixes []IPv6Entry `json:"ipv6Prefixes" yaml:"ipv6Prefixes"`
}
