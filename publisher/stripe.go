package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/stripe"
)

const stripeFile = "stripe.json"

func fetchStripe() ([]byte, error) {
	a := stripe.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncStripeData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(stripeFile, data, wt, fs)
}
