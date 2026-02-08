package github

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
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
	Timeout     time.Duration
}

func New() GitHub {
	return GitHub{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (gh *GitHub) FetchData() ([]byte, http.Header, int, error) {
	if gh.DownloadURL == "" {
		gh.DownloadURL = DownloadURL
	}

	return web.Request(gh.Client, gh.DownloadURL, http.MethodGet, nil, nil, gh.Timeout)
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
