// Package binarydefense retrieves Binary Defense's public artillery banlist.
package binarydefense

import (
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "binarydefense"
	FullName  = "Binary Defense"
	HostType  = "threat"
	SourceURL = "https://www.binarydefense.com/"
	// DownloadURL returns hosts Binary Defense's artillery sensors have
	// observed attacking, as a newline separated list of bare addresses behind
	// a commented header.
	DownloadURL = "https://www.binarydefense.com/banlist.txt"
)

type BinaryDefense struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() BinaryDefense {
	return BinaryDefense{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (b *BinaryDefense) FetchData() ([]byte, http.Header, int, error) {
	if b.DownloadURL == "" {
		b.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(b.Client, b.DownloadURL, http.MethodGet, nil, nil, b.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download binary defense banlist from %s. http status code: %d", b.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (b *BinaryDefense) Fetch() (Doc, error) {
	data, _, _, err := b.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	ipv4, ipv6, err := iplist.Parse(ShortName, data)
	if err != nil {
		return Doc{}, err
	}

	return Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}, nil
}
