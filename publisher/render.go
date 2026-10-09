package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/render"
)

const renderFile = "render.json"

func fetchRender() ([]byte, error) {
	a := render.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncRenderData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(renderFile, data, wt, fs)
}
