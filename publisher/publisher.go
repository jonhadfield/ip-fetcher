package publisher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/sync/errgroup"
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

		os.Exit(1)
	}
}

func New() *Publisher {
	var pub Publisher

	pub.GitHubRepoURL = strings.TrimSpace(os.Getenv("GITHUB_PUBLISH_URL"))
	if pub.GitHubRepoURL == "" {
		slog.Error("GITHUB_PUBLISH_URL not set") //nolint:sloglint
		os.Exit(1)
	}

	pub.GitHubToken = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if pub.GitHubToken == "" {
		slog.Error("GITHUB_TOKEN not set") //nolint:sloglint
		os.Exit(1)
	}

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
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Phase 1: Fetch all provider data in parallel
	type fetchResult struct {
		data []byte
		err  error
	}

	results := make([]fetchResult, len(providers))

	var g errgroup.Group

	for i, provider := range providers {
		g.Go(func() error {
			data, fetchErr := provider.FetchFunc()
			results[i] = fetchResult{data: data, err: fetchErr}

			return nil // don't fail fast — collect all results
		})
	}

	_ = g.Wait()

	// Phase 2: Sync sequentially (git operations are not concurrency-safe)
	var included []string

	for i, provider := range providers {
		if results[i].err != nil {
			slog.Info("failed to fetch", "provider", provider.ShortName, "error", results[i].err)

			continue
		}

		var commit plumbing.Hash

		commit, err = provider.SyncDataFunc(results[i].data, w, fs)
		if err != nil {
			slog.Info("failed to sync", "provider", provider.ShortName, "error", err)

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

	return hex.EncodeToString(hash.Sum(nil)), nil
}
