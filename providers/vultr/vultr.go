package vultr

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "vultr"
	FullName    = "Vultr"
	HostType    = "hosting"
	SourceURL   = "https://www.vultr.com/"
	DownloadURL = bgpview.DefaultURL
)

var ASNs = []string{"20473"} //nolint:nolintlint,gochecknoglobals

type Vultr struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() Vultr {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Vultr{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      c,
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (h *Vultr) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName, h.Timeout)
}

func (h *Vultr) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
