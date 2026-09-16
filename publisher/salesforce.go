package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/salesforce"
)

const salesforceFile = "salesforce.json"

func fetchSalesforce() ([]byte, error) {
	a := salesforce.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncSalesforceData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(salesforceFile, data, wt, fs)
}
