package publisher

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"

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
	rgb, err := fs.Open(cloudflareFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(cloudflareFile, "up to date", upToDate)
	}

	if err = createFile(fs, cloudflareFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(cloudflareFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update cloudflare data")
}
