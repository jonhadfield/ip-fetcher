package hetzner

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "hetzner"
	FullName    = "Hetzner"
	HostType    = "hosting"
	SourceURL   = "https://www.hetzner.com/"
	DownloadURL = bgpview.DefaultURL
)

var ASNs = []string{"24940", "213230", "212317", "215859"} //nolint:nolintlint,gochecknoglobals

type Hetzner struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
}

type Doc = bgpview.Doc

func New() Hetzner {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Hetzner{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      c,
	}
}

func (h *Hetzner) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName)
}

func (h *Hetzner) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
