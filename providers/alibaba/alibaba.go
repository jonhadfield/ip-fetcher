package alibaba

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
)

const (
	ShortName   = "alibaba"
	FullName    = "Alibaba"
	HostType    = "hosting"
	SourceURL   = "https://www.alibaba.com/"
	DownloadURL = bgpview.DefaultURL
)

var ASNs = []string{"45102", "24429", "134963", "37963"} //nolint:nolintlint,gochecknoglobals

type Alibaba struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() Alibaba {
	return Alibaba{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (h *Alibaba) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName, h.Timeout)
}

func (h *Alibaba) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
