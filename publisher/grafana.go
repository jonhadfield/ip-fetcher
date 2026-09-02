package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/grafana"
)

const grafanaFile = "grafana.json"

func fetchGrafana() ([]byte, error) {
	a := grafana.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGrafanaData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(grafanaFile, data, wt, fs)
}
