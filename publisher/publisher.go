package publisher

import (
	"crypto/sha256"
	"fmt"
	"github.com/jonhadfield/ip-fetcher/providers/aws"
	"github.com/jonhadfield/ip-fetcher/providers/azure"
	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
	"github.com/jonhadfield/ip-fetcher/providers/gcp"
	"io"
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

const (
	sAWS        = "aws"
	sAzure      = "azure"
	sCloudflare = "cloudflare"
	sFastly     = "fastly"
	sGCP        = "gcp"
	sGoogle     = "google"
	sGooglebot  = "googlebot"
	sGoogleSC   = "googlesc"
	sGoogleUTF  = "googleutf"
	sLinode     = "linode"
	sOCI        = "oci"
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

var providerList = []string{
	sAWS,
	sAzure,
	sCloudflare,
	sFastly,
	sGCP,
	sGoogle,
	sGooglebot,
	sGoogleSC,
	sGoogleUTF,
	sLinode,
	sOCI,
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

	for _, provider := range providerList {
		var commit plumbing.Hash

		// TODO: allow stale if source is down
		switch provider {
		case aws.ShortName:
			commit, err = syncAWS(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case azure.ShortName:
			commit, err = syncAzure(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case cloudflare.ShortName:
			commit, err = syncCloudflare(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sFastly:
			commit, err = syncFastly(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case gcp.ShortName:
			commit, err = syncGCP(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sGoogle:
			commit, err = syncGoogle(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sGooglebot:
			commit, err = syncGooglebot(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sGoogleSC:
			commit, err = syncGoogleSC(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sGoogleUTF:
			commit, err = syncGoogleUTF(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sLinode:
			commit, err = syncLinode(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
		case sOCI:
			commit, err = syncOCI(w, fs)
			if err != nil {
				slog.Error("failed to sync", provider, err.Error())

				break
			}

			included = append(included, provider)
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
