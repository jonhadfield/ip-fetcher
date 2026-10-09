package publisher

import (
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
	return syncData(perplexitybotFile, data, wt, fs)
}
