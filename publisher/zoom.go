package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/zoom"
)

const zoomFile = "zoom.txt"

func fetchZoom() ([]byte, error) {
	a := zoom.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncZoomData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(zoomFile, data, wt, fs)
}
