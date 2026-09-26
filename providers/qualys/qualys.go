// Package qualys retrieves prefixes announced by Qualys (AS27385).
//
// Qualys does not publish a single public scanner allowlist: the SOC ranges
// shown under Help > About are account-specific. The prefixes announced by
// AS27385 are Qualys-owned space (including the well known 64.39.x and 69.67.x
// blocks), so an address that matches this document is Qualys. Addresses that
// do not match may still be Qualys scanners hosted in a platform SOC.
package qualys

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
)

const (
	ShortName   = "qualys"
	FullName    = "Qualys"
	HostType    = "scanner"
	SourceURL   = "https://docs.qualys.com/en/vm/latest/help/external_scanner_ips.htm"
	DownloadURL = bgpview.DefaultURL
)

// ASNs is Qualys, Inc. (AS27385).
var ASNs = []string{"27385"} //nolint:nolintlint,gochecknoglobals

type Qualys struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() Qualys {
	return Qualys{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.LongRequestTimeout,
	}
}

func (q *Qualys) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(q.Client, q.DownloadURL, q.ASNs, FullName, q.Timeout)
}

func (q *Qualys) Fetch() (Doc, error) {
	data, _, _, err := q.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
