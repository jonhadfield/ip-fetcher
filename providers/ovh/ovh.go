package ovh

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "ovh"
	FullName    = "OVH"
	HostType    = "hosting"
	SourceURL   = "https://www.ovh.com/"
	DownloadURL = bgpview.DefaultURL
)

var ASNs = []string{"16276"} //nolint:nolintlint,gochecknoglobals

type OVH struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
}

type Doc = bgpview.Doc

func New() OVH {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return OVH{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      c,
	}
}

func (h *OVH) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName)
}

func (h *OVH) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
