package publisher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	sAWS       = "aws"
	sAzure     = "azure"
	sGCP       = "gcp"
	sGoogle    = "google"
	sGooglebot = "googlebot"
	sLinode    = "linode"
)

type Publisher struct {
	GitHubToken   string
	GitHubRepoURL string
}

func Publish() {
	p := New()

	err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func New() *Publisher {
	var pub Publisher

	pub.GitHubRepoURL = os.Getenv("GITHUB_PUBLISH_URL")
	pub.GitHubToken = os.Getenv("GITHUB_TOKEN")

	return &pub
}

var providerList = []string{sAWS, sAzure, sGCP, sGoogle, sGooglebot, sLinode}

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

	// for each provider, fetch latest data and compare with the current file in the repo
	var needPush bool

	for _, provider := range providerList {
		var commit plumbing.Hash

		switch provider {
		case sAWS:
			commit, err = syncAWSData(w, fs)
		case sAzure:
			commit, err = syncAzureData(w, fs)
		case sGCP:
			commit, err = syncGCP(w, fs)
		case sGooglebot:
			commit, err = syncGooglebot(w, fs)
		case sGoogle:
			commit, err = syncGoogle(w, fs)
		case sLinode:
			commit, err = syncLinode(w, fs)
		}

		if err != nil {
			slog.Error("failed to sync", provider, err.Error())

			return err
		}

		if commit.IsZero() {
			slog.Info("provider", provider, "in sync")

			continue
		}

		slog.Info("provider", provider, "not in sync")

		_, err = repo.CommitObject(commit)
		if err != nil {
			return err
		}

		needPush = true
	}

	if !needPush {
		return nil
	}

	slog.Info("pushing changes")

	err = repo.Push(&git.PushOptions{Auth: &http.BasicAuth{
		Username: "ip-fetcher",
		Password: p.GitHubToken,
	}})
	if err != nil {
		return err
	}

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

	var n int

	n, err = gbFile.Write(content)
	if err != nil {
		return err
	}

	slog.Info("wrote bytes", strconv.Itoa(n), name)

	return nil
}

func fileContentHash(content io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, content); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
