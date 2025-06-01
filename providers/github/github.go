package github

import (
	"encoding/json"
	"net/http"
	"net/netip"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "github"
	FullName    = "GitHub"
	HostType    = "hosting"
	SourceURL   = "https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-githubs-ip-addresses"
	DownloadURL = "https://api.github.com/meta"
)

type GitHub struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() GitHub {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return GitHub{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

func (gh *GitHub) FetchData() ([]byte, http.Header, int, error) {
	if gh.DownloadURL == "" {
		gh.DownloadURL = DownloadURL
	}

	return web.Request(gh.Client, gh.DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (gh *GitHub) Fetch() ([]netip.Prefix, error) {
	data, _, _, err := gh.FetchData()
	if err != nil {
		return nil, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) ([]netip.Prefix, error) {
	var raw map[string]json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var prefixes []netip.Prefix

	for _, v := range raw {
		var entries []string
		if err := json.Unmarshal(v, &entries); err != nil {
			continue
		}

		for _, entry := range entries {
			prefix, perr := netip.ParsePrefix(entry)
			if perr != nil {
				continue
			}

			prefixes = append(prefixes, prefix)
		}
	}

	return prefixes, nil
}
