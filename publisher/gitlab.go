package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/gitlab"
)

const gitlabFile = "gitlab.txt"

func fetchGitLab() ([]byte, error) {
	a := gitlab.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGitLabData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(gitlabFile, data, wt, fs)
}
