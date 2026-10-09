package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/datadog"
)

const datadogFile = "datadog.json"

func fetchDatadog() ([]byte, error) {
	a := datadog.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncDatadogData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(datadogFile, data, wt, fs)
}
