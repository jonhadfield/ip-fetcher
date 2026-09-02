package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/sentry"
)

const sentryFile = "sentry.txt"

func fetchSentry() ([]byte, error) {
	a := sentry.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncSentryData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(sentryFile, data, wt, fs)
}
