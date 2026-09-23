// Package circleci retrieves CircleCI's published IP ranges for job egress.
package circleci

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
	ShortName = "circleci"
	FullName  = "CircleCI"
	HostType  = "saas"
	SourceURL = "https://circleci.com/docs/ip-ranges/"
	// DownloadURL is CircleCI's machine-readable list of IP ranges used when
	// the IP ranges feature is enabled for a job.
	DownloadURL = "https://circleci.com/docs/ip-ranges-list.json"
)

type CircleCI struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() CircleCI {
	return CircleCI{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawDoc struct {
	IPRanges RawRanges `json:"IPRanges"`
}

type RawRanges struct {
	MacOS []string `json:"macOS"`
	Jobs  []string `json:"jobs"`
	Core  []string `json:"core"`
}

type Doc struct {
	MacOS        []netip.Prefix `json:"macos"         yaml:"macos"`
	Jobs         []netip.Prefix `json:"jobs"          yaml:"jobs"`
	Core         []netip.Prefix `json:"core"          yaml:"core"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (c *CircleCI) FetchData() ([]byte, http.Header, int, error) {
	if c.DownloadURL == "" {
		c.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(c.Client, c.DownloadURL, http.MethodGet, nil, nil, c.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download circleci prefixes from %s. http status code: %d", c.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (c *CircleCI) Fetch() (Doc, error) {
	data, _, _, err := c.FetchData()
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

	doc := Doc{
		MacOS: parseEntries(raw.IPRanges.MacOS),
		Jobs:  parseEntries(raw.IPRanges.Jobs),
		Core:  parseEntries(raw.IPRanges.Core),
	}

	for _, group := range [][]netip.Prefix{doc.MacOS, doc.Jobs, doc.Core} {
		for _, prefix := range group {
			if prefix.Addr().Is4() {
				doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

				continue
			}

			doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
		}
	}

	return doc, nil
}

func parseEntries(entries []string) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, entry := range entries {
		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse circleci address: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes
}
