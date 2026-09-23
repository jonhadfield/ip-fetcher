package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/telegram"
)

const telegramFile = "telegram.txt"

func fetchTelegram() ([]byte, error) {
	a := telegram.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncTelegramData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(telegramFile, data, wt, fs)
}
