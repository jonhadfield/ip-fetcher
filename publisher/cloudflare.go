package publisher

import (
	"encoding/json"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
)

const cloudflareFile = "cloudflare.json"

func fetchCloudflare() ([]byte, error) {
	a := cloudflare.New()

	prefixes, err := a.Fetch()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(prefixes, "", "  ")
}

func syncCloudflareData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(cloudflareFile, data, wt, fs)
}
