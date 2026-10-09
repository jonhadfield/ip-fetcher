package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/oci"
)

const ociFile = "oci.json"

func fetchOCI() ([]byte, error) {
	a := oci.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncOCIData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ociFile, data, wt, fs)
}
