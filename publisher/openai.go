package publisher

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/openai"
)

const openaiFile = "openai.json"

// fetchOpenAI combines the per-bot feeds into the single document the CLI
// writes, as OpenAI publishes GPTBot, OAI-SearchBot and ChatGPT-User separately.
func fetchOpenAI() ([]byte, error) {
	o := openai.New()

	doc, err := o.Fetch()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(doc, "", "  ")
}

func syncOpenAIData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(openaiFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(openaiFile, "up to date", upToDate)
	}

	if err = createFile(fs, openaiFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(openaiFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update openai data")
}
