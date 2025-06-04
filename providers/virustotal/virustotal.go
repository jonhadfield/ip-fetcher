package virustotal

import (
	"bufio"
	"bytes"
	"net/http"
	"net/netip"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "virustotal"
	FullName    = "VirusTotal"
	HostType    = "security"
	SourceURL   = "https://support.virustotal.com/"
	DownloadURL = "https://www.virustotal.com/static/ip-addresses.txt"
)

type VirusTotal struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() VirusTotal {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return VirusTotal{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

func (v *VirusTotal) FetchData() ([]byte, http.Header, int, error) {
	if v.DownloadURL == "" {
		v.DownloadURL = DownloadURL
	}

	return web.Request(v.Client, v.DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (v *VirusTotal) Fetch() ([]netip.Prefix, error) {
	data, _, _, err := v.FetchData()
	if err != nil {
		return nil, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) ([]netip.Prefix, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var prefixes []netip.Prefix

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}
