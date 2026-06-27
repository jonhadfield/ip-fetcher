package leaseweb

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
)

const (
	ShortName   = "leaseweb"
	FullName    = "Leaseweb"
	HostType    = "hosting"
	SourceURL   = "https://www.leaseweb.com/"
	DownloadURL = bgpview.DefaultURL
)

// ASNs covers Leaseweb's primary regional autonomous systems (NL, DE and US).
// Leaseweb announces prefixes under several ASNs; these carry the bulk of them.
var ASNs = []string{"60781", "28753", "30633", "395954"} //nolint:nolintlint,gochecknoglobals

type Leaseweb struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() Leaseweb {
	return Leaseweb{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.LongRequestTimeout,
	}
}

func (h *Leaseweb) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName, h.Timeout)
}

func (h *Leaseweb) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
