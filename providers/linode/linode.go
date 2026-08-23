package linode

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/geocsv"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName   = "linode"
	FullName    = "Linode"
	HostType    = "hosting"
	SourceURL   = "https://www.linode.com/"
	DownloadURL = "https://geoip.linode.com/"
)

// The document format is identical to icloudpr's, so the parsing lives in
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

type Linode struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Linode {
	return Linode{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (a *Linode) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return geocsv.FetchData(a.Client, a.DownloadURL, a.Timeout)
}

func (a *Linode) Fetch() (Doc, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return geocsv.Fetch(a.Client, a.DownloadURL, a.Timeout)
}
