package googlesc

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
	ShortName                = "googlesc"
	FullName                 = "Google Special Crawlers"
	HostType                 = "crawlers"
	SourceURL                = "https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot"
	DownloadURL              = "https://developers.google.com/static/search/apis/ipranges/special-crawlers.json"
	downloadedFileTimeFormat = "2006-01-02T15:04:05.999999"
)

func New() Googlesc {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Googlesc{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

type Googlesc struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

type RawDoc struct {
	CreationTime  string `json:"creationTime"`
	LastRequested time.Time
	Entries       []json.RawMessage `json:"prefixes"`
}

func (gs *Googlesc) FetchData() ([]byte, http.Header, int, error) {
	var (
		data    []byte
		headers http.Header
		status  int
		err     error
	)
	if gs.DownloadURL == "" {
		gs.DownloadURL = DownloadURL
	}
	data, headers, status, err = web.Request(
		gs.Client,
		gs.DownloadURL,
		http.MethodGet,
		nil,
		nil,
		web.DefaultRequestTimeout,
	)
	return data, headers, status, err
}

func (gs *Googlesc) Fetch() (Doc, error) {
	data, _, _, err := gs.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var (
		doc    Doc
		rawDoc RawDoc
	)
	err := json.Unmarshal(data, &rawDoc)
	if err != nil {
		return Doc{}, err
	}

	doc.IPv4Prefixes, doc.IPv6Prefixes, err = castEntries(rawDoc.Entries)
	if err != nil {
		return Doc{}, err
	}

	ct, err := time.Parse(downloadedFileTimeFormat, rawDoc.CreationTime)
	if err != nil {
		return Doc{}, err
	}

	doc.CreationTime = ct

	return doc, nil
}

func castEntries(prefixes []json.RawMessage) ([]IPv4Entry, []IPv6Entry, error) {
	var (
		ipv4 []IPv4Entry
		ipv6 []IPv6Entry
		err  error
	)
	for _, pr := range prefixes {
		var ipv4entry RawIPv4Entry

		var ipv6entry RawIPv6Entry

		// try 4
		err = json.Unmarshal(pr, &ipv4entry)
		if err == nil {
			ipv4Prefix, parseError := netip.ParsePrefix(ipv4entry.IPv4Prefix)
			if parseError == nil {
				ipv4 = append(ipv4, IPv4Entry{
					IPv4Prefix: ipv4Prefix,
				})

				continue
			}
		}

		// try 6
		err = json.Unmarshal(pr, &ipv6entry)
		if err == nil {
			ipv6Prefix, parseError := netip.ParsePrefix(ipv6entry.IPv6Prefix)
			if parseError != nil {
				return ipv4, ipv6, parseError
			}

			ipv6 = append(ipv6, IPv6Entry{
				IPv6Prefix: ipv6Prefix,
			})

			continue
		}

		if err != nil {
			return ipv4, ipv6, err
		}
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
	CreationTime time.Time
	IPv4Prefixes []IPv4Entry
	IPv6Prefixes []IPv6Entry
}
