package m247

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
	"github.com/sirupsen/logrus"
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
}

type Doc = bgpview.Doc

func New() M247 {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return M247{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      c,
	}
}

func (h *M247) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName)
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
