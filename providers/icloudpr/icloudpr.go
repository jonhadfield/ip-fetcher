package icloudpr

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/geocsv"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "icloudpr"
	FullName    = "iCloud Private Relay"
	HostType    = "anonymiser"
	SourceURL   = "https://support.apple.com/en-us/HT212614"
	DownloadURL = "https://mask-api.icloud.com/egress-ip-ranges.csv"
)

// The document format is identical to linode's, so the parsing lives in
// internal/geocsv. These are aliases, not new types, so this package's
// API is unchanged.

type (
	Doc      = geocsv.Doc
	Record   = geocsv.Record
	Entry    = geocsv.Entry
	CSVEntry = geocsv.CSVEntry
)

func IsIPv4(address string) bool { return geocsv.IsIPv4(address) }

func IsIPv6(address string) bool { return geocsv.IsIPv6(address) }

func Parse(data []byte) ([]Record, error) { return geocsv.Parse(data) }

type ICloudPrivateRelay struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() ICloudPrivateRelay {
	return ICloudPrivateRelay{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (a *ICloudPrivateRelay) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return geocsv.FetchData(a.Client, a.DownloadURL, a.Timeout)
}

func (a *ICloudPrivateRelay) Fetch() (Doc, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return geocsv.Fetch(a.Client, a.DownloadURL, a.Timeout)
}
