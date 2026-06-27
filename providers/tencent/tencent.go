package tencent

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
)

const (
	ShortName   = "tencent"
	FullName    = "Tencent Cloud"
	HostType    = "hosting"
	SourceURL   = "https://www.tencentcloud.com/"
	DownloadURL = bgpview.DefaultURL
)

// ASNs covers Tencent's primary global cloud autonomous systems. Tencent
// announces prefixes under several ASNs; these two carry the bulk of the
// Tencent Cloud ranges.
var ASNs = []string{"132203", "45090"} //nolint:nolintlint,gochecknoglobals

type Tencent struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() Tencent {
	return Tencent{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.LongRequestTimeout,
	}
}

func (h *Tencent) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName, h.Timeout)
}

func (h *Tencent) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
