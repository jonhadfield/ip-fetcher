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

const (
	cloudflareFile      = "cloudflare.json"
	errFailedToDownload = "failed to retrieve cloudflare prefixes initial page"
)

func syncCloudflare(wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	a := cloudflare.New()

	originContent, err := a.Fetch()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	mContent, err := json.MarshalIndent(originContent, "", "  ")
	if err != nil {
		return plumbing.ZeroHash, err
	}

	rgb, err := fs.Open(cloudflareFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}
	// if the file doesn't exist, we need to create it
	if err == nil {
		var upToDate bool

		upToDate, err = isUpToDate(bytes.NewReader(mContent), rgb)
		if err != nil || upToDate {
			return plumbing.ZeroHash, err
		}

		slog.Info(cloudflareFile, "up to date", upToDate)
	}

	if err = createFile(fs, cloudflareFile, mContent); err != nil {
		return plumbing.ZeroHash, err
	}

	// Adds the new file to the staging area.
	_, err = wt.Add(cloudflareFile)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update cloudflare data")
}
