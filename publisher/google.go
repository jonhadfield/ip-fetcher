package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/google"
)

const googleFile = "google.json"

func fetchGoogle() ([]byte, error) {
	a := google.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGoogleData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(googleFile, data, wt, fs)
}
