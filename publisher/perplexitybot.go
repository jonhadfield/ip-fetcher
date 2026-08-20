package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/perplexitybot"
)

const perplexitybotFile = "perplexitybot.json"

func fetchPerplexitybot() ([]byte, error) {
	a := perplexitybot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncPerplexitybotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(perplexitybotFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(perplexitybotFile, "up to date", upToDate)
	}

	if err = createFile(fs, perplexitybotFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(perplexitybotFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update perplexitybot data")
}
