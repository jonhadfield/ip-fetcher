package huawei

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/bgpview"
)

const (
	ShortName   = "huawei"
	FullName    = "Huawei Cloud"
	HostType    = "hosting"
	SourceURL   = "https://www.huaweicloud.com/"
	DownloadURL = bgpview.DefaultURL
)

// ASNs covers Huawei Cloud's primary autonomous systems as published in RADB
// AS-HUAWEI.
var ASNs = []string{"131444", "136907", "141180", "206204", "206798", "265443", "55990"} //nolint:nolintlint,gochecknoglobals

type Huawei struct {
	Client      *retryablehttp.Client
	DownloadURL string
	ASNs        []string
	Timeout     time.Duration
}

type Doc = bgpview.Doc

func New() Huawei {
	return Huawei{
		DownloadURL: bgpview.DefaultURL,
		ASNs:        ASNs,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.LongRequestTimeout,
	}
}

func (h *Huawei) FetchData() ([]byte, http.Header, int, error) {
	return bgpview.FetchData(h.Client, h.DownloadURL, h.ASNs, FullName, h.Timeout)
}

func (h *Huawei) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	return bgpview.ProcessData(data, FullName)
}
