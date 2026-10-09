package publisher

import (
	"encoding/json"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/openai"
)

const openaiFile = "openai.json"

// fetchOpenAI combines the per-bot feeds into the single document the CLI
// writes, as OpenAI publishes GPTBot, OAI-SearchBot, ChatGPT-User and
// OAI-AdsBot separately.
func fetchOpenAI() ([]byte, error) {
	o := openai.New()

	doc, err := o.Fetch()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(doc, "", "  ")
}

func syncOpenAIData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(openaiFile, data, wt, fs)
}
