package publisher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Publisher struct {
	GitHubToken   string
	GitHubRepoURL string
}

func Publish() {
	p := New()

	err := p.Run()
	if err != nil {
		slog.Error("publish failed", "error", err)
	}
}

func New() *Publisher {
	var pub Publisher

	pub.GitHubRepoURL = os.Getenv("GITHUB_PUBLISH_URL")
	pub.GitHubToken = os.Getenv("GITHUB_TOKEN")

	return &pub
}

func (p *Publisher) Run() error {
	fs := memfs.New()
	storer := memory.NewStorage()

	repo, err := git.Clone(storer, fs, &git.CloneOptions{Auth: &http.BasicAuth{
		Username: "-",
		Password: p.GitHubToken,
	}, URL: p.GitHubRepoURL})
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// create list of providers to include in README
	var included []string

	for _, provider := range providers {
		var commit plumbing.Hash

		commit, err = provider.SyncFunc(w, fs)
		if err != nil {
			log.Printf("failed to sync %s: %v", provider.ShortName, err)
			continue
		}

		included = append(included, provider.ShortName)

		if commit.IsZero() {
			slog.Info("provider", provider.ShortName, "in sync")

			continue
		}

		slog.Info("provider", provider.ShortName, "not in sync")

		_, err = repo.CommitObject(commit)
		if err != nil {
			return err
		}
	}

	commit, err := syncReadMe(included, w, fs)
	if err != nil {
		return err
	}

	_, err = repo.CommitObject(commit)
	if err != nil {
		return err
	}

	slog.Info("pushing changes")

	err = repo.Push(&git.PushOptions{Auth: &http.BasicAuth{
		Username: "ip-fetcher",
		Password: p.GitHubToken,
	}})
	if err != nil {
		return err
	}

	slog.Info("publish successful", "url", p.GitHubRepoURL)

	return nil
}

func isUpToDate(origin, repo io.Reader) (bool, error) {
	originHash, err := fileContentHash(origin)
	if err != nil {
		return false, err
	}

	var repoHash string

	repoHash, err = fileContentHash(repo)
	if err != nil {
		return false, err
	}

	if originHash == repoHash {
		return true, nil
	}

	return false, nil
}

func createCommit(wt *git.Worktree, msg string) (plumbing.Hash, error) {
	var err error

	var commit plumbing.Hash

	commit, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "ip-fetcher",
			Email: "ip-fetcher@lessknown.co.uk",
			When:  time.Now(),
		},
	})
	if err != nil {
		return plumbing.Hash{}, err
	}

	return commit, nil
}

func createFile(fs billy.Filesystem, name string, content []byte) error {
	var err error

	var gbFile billy.File

	gbFile, err = fs.Create(name)
	if err != nil {
		return err
	}

	if _, err = gbFile.Write(content); err != nil {
		return err
	}

	if name == "README.md" {
		slog.Info("README.md updated")

		return nil
	}

	slog.Info("provider", name, "updated")

	return nil
}

func fileContentHash(content io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, content); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
