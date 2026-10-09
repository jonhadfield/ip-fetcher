package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/providers/akamai"
)

const akamaiFile = "akamai.txt"

// fetchAkamai returns the parsed prefixes rather than the raw response, as
// Akamai publishes its CIDR lists inside a zip archive.
func fetchAkamai() ([]byte, error) {
	a := akamai.New()

	prefixes, err := a.Fetch()
	if err != nil {
		return nil, err
	}

	return iplist.ToLines(prefixes), nil
}

func syncAkamaiData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(akamaiFile, data, wt, fs)
}
