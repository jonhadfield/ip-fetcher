// Package onetrust retrieves OneTrust web scanner egress prefixes.
package onetrust

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
	ShortName   = "onetrust"
	FullName    = "OneTrust"
	HostType    = "scanner"
	SourceURL   = "https://my.onetrust.com/s/article/UUID-21f6bff2-1b12-8c67-e8b0-d852e36f37af"
	DownloadURL = "https://my.onetrust.com/s/sfsites/c/resource/platform/hosting/webscanner-ips.txt"
)

type OneTrust struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() OneTrust {
	return OneTrust{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (o *OneTrust) FetchData() ([]byte, http.Header, int, error) {
	if o.DownloadURL == "" {
		o.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(o.Client, o.DownloadURL, http.MethodGet, nil, nil, o.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download onetrust addresses from %s. http status code: %d", o.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (o *OneTrust) Fetch() (Doc, error) {
	data, _, _, err := o.FetchData()
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
