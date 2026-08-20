package anthropic

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "anthropic"
	FullName    = "Anthropic Crawler Bots"
	HostType    = "crawlers"
	SourceURL   = "https://claude.com/crawling/bots.json"
	DownloadURL = "https://claude.com/crawling/bots.json"
)

// creationTime is published as RFC3339, but sibling crawler feeds use a
// fractional seconds variant, so both are accepted.
var creationTimeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05.999999",
}

func New() Anthropic {
	return Anthropic{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Anthropic struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

type RawDoc struct {
	CreationTime  string            `json:"creationTime"`
	LastRequested time.Time         `json:"-" yaml:"-"`
	Entries       []json.RawMessage `json:"prefixes"`
}

func (a *Anthropic) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return web.Request(
		a.Client,
		a.DownloadURL,
		http.MethodGet,
		nil,
		nil,
		a.Timeout,
	)
}

func (a *Anthropic) Fetch() (Doc, error) {
	data, _, _, err := a.FetchData()
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

	// creationTime may be absent from the feed, so treat it as optional.
	if rawDoc.CreationTime != "" {
		var ct time.Time

		ct, err = parseCreationTime(rawDoc.CreationTime)
		if err != nil {
			return Doc{}, err
		}

		doc.CreationTime = ct
	}

	return doc, nil
}

func parseCreationTime(in string) (time.Time, error) {
	var lastErr error

	for _, format := range creationTimeFormats {
		ct, err := time.Parse(format, in)
		if err == nil {
			return ct, nil
		}

		lastErr = err
	}

	return time.Time{}, lastErr
}

func castEntries(prefixes []json.RawMessage) ([]IPv4Entry, []IPv6Entry, error) {
	var (
		ipv4 []IPv4Entry
		ipv6 []IPv6Entry
	)

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
