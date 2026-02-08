package m247

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
)

const (
	ShortName   = "m247"
	FullName    = "M247"
	HostType    = "hosting"
	SourceURL   = "https://www.m247.com/"
	DownloadURL = bgpview.DefaultURL
)

var ASNs = []string{"16247"} //nolint:nolintlint,gochecknoglobals

type M247 struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() M247 {
	return M247{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (h *M247) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName, h.Timeout)
}

func (h *M247) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
