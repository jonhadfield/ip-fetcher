// Package threatfox retrieves recent ip:port indicators from abuse.ch ThreatFox.
package threatfox

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "threatfox"
	FullName  = "abuse.ch ThreatFox"
	HostType  = "threat"
	SourceURL = "https://threatfox.abuse.ch/"
	// DownloadURL returns recent ip:port indicators as a JSON object keyed by
	// IOC id. Hosts are extracted for prefix output; the upstream document is
	// what the publisher commits so malware metadata is preserved.
	DownloadURL = "https://threatfox.abuse.ch/export/json/ip-port/recent/"
)

type ThreatFox struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() ThreatFox {
	return ThreatFox{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// RawIOC is one indicator in the upstream export.
type RawIOC struct {
	IOCValue         string `json:"ioc_value"`
	IOCType          string `json:"ioc_type"`
	ThreatType       string `json:"threat_type"`
	Malware          string `json:"fk_malware"`
	MalwarePrintable string `json:"malware_printable"`
	ConfidenceLevel  int    `json:"confidence_level"`
	FirstSeenUTC     string `json:"first_seen_utc"`
}

// Entry is a parsed indicator with its host as a prefix.
type Entry struct {
	Prefix           netip.Prefix `json:"prefix"            yaml:"prefix"`
	Port             string       `json:"port,omitempty"    yaml:"port,omitempty"`
	ThreatType       string       `json:"threat_type"       yaml:"threat_type"`
	Malware          string       `json:"malware"           yaml:"malware"`
	MalwarePrintable string       `json:"malware_printable" yaml:"malware_printable"`
	ConfidenceLevel  int          `json:"confidence_level"  yaml:"confidence_level"`
	FirstSeenUTC     string       `json:"first_seen_utc"    yaml:"first_seen_utc"`
}

type Doc struct {
	Entries      []Entry        `json:"entries"       yaml:"entries"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (t *ThreatFox) FetchData() ([]byte, http.Header, int, error) {
	if t.DownloadURL == "" {
		t.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(t.Client, t.DownloadURL, http.MethodGet, nil, nil, t.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download threatfox indicators from %s. http status code: %d", t.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (t *ThreatFox) Fetch() (Doc, error) {
	data, _, _, err := t.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw map[string][]RawIOC
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	var (
		doc  Doc
		seen = make(map[netip.Prefix]struct{})
	)

	for _, iocs := range raw {
		for _, ioc := range iocs {
			host, port, ok := splitHostPort(ioc.IOCValue)
			if !ok {
				logrus.Warnf("failed to parse threatfox ioc: %s", ioc.IOCValue)

				continue
			}

			prefix, ok := iplist.ToPrefix(host)
			if !ok {
				logrus.Warnf("failed to parse threatfox host: %s", host)

				continue
			}

			doc.Entries = append(doc.Entries, Entry{
				Prefix:           prefix,
				Port:             port,
				ThreatType:       ioc.ThreatType,
				Malware:          ioc.Malware,
				MalwarePrintable: ioc.MalwarePrintable,
				ConfidenceLevel:  ioc.ConfidenceLevel,
				FirstSeenUTC:     ioc.FirstSeenUTC,
			})

			if _, exists := seen[prefix]; exists {
				continue
			}

			seen[prefix] = struct{}{}

			if prefix.Addr().Is4() {
				doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

				continue
			}

			doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
		}
	}

	slices.SortFunc(doc.Entries, func(a, b Entry) int {
		if c := a.Prefix.Addr().Compare(b.Prefix.Addr()); c != 0 {
			return c
		}

		return cmp.Compare(a.Port, b.Port)
	})
	slices.SortFunc(doc.IPv4Prefixes, func(a, b netip.Prefix) int {
		return a.Addr().Compare(b.Addr())
	})
	slices.SortFunc(doc.IPv6Prefixes, func(a, b netip.Prefix) int {
		return a.Addr().Compare(b.Addr())
	})

	return doc, nil
}

func splitHostPort(value string) (string, string, bool) {
	if _, err := netip.ParseAddr(value); err == nil {
		return value, "", true
	}

	h, p, err := net.SplitHostPort(value)
	if err != nil {
		return "", "", false
	}

	return h, p, true
}
