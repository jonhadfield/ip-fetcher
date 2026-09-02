// the sync helpers are unexported, so these tests live in the package itself.
package publisher //nolint:testpackage

import (
	"io"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

func newTestWorktree(t *testing.T) (*git.Worktree, billy.Filesystem) {
	t.Helper()

	fs := memfs.New()

	repo, err := git.Init(memory.NewStorage(), fs)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	return wt, fs
}

func readFile(t *testing.T, fs billy.Filesystem, name string) string {
	t.Helper()

	f, err := fs.Open(name)
	require.NoError(t, err)

	content, err := io.ReadAll(f)
	require.NoError(t, err)

	return string(content)
}

// a document the repository does not hold yet is written and committed.
func TestSyncDataCommitsNewFile(t *testing.T) {
	wt, fs := newTestWorktree(t)

	commit, err := syncData("grafana.json", []byte("first\n"), wt, fs)
	require.NoError(t, err)
	require.NotEqual(t, plumbing.ZeroHash, commit)
	require.Equal(t, "first\n", readFile(t, fs, "grafana.json"))
}

// an unchanged document produces no second commit.
func TestSyncDataSkipsUnchangedFile(t *testing.T) {
	wt, fs := newTestWorktree(t)

	data := []byte("34.123.33.225\n")

	first, err := syncData("sentry.txt", data, wt, fs)
	require.NoError(t, err)
	require.NotEqual(t, plumbing.ZeroHash, first)

	second, err := syncData("sentry.txt", data, wt, fs)
	require.NoError(t, err)
	require.Equal(t, plumbing.ZeroHash, second)
}

// a changed document is written over the old one and committed again.
func TestSyncDataCommitsChangedFile(t *testing.T) {
	wt, fs := newTestWorktree(t)

	first, err := syncData("updown.json", []byte("one node\n"), wt, fs)
	require.NoError(t, err)

	second, err := syncData("updown.json", []byte("two nodes\n"), wt, fs)
	require.NoError(t, err)
	require.NotEqual(t, plumbing.ZeroHash, second)
	require.NotEqual(t, first, second)
	require.Equal(t, "two nodes\n", readFile(t, fs, "updown.json"))
}

// each provider's sync function writes to the file the readme rows name.
func TestSyncProviderDataWritesProviderFile(t *testing.T) {
	cases := []struct {
		name string
		sync func(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error)
		file string
	}{
		{"grafana", syncGrafanaData, grafanaFile},
		{"sentry", syncSentryData, sentryFile},
		{"site24x7", syncSite24x7Data, site24x7File},
		{"updown", syncUpdownData, updownFile},
		{"uptrends", syncUptrendsData, uptrendsFile},
		{"tenable", syncTenableData, tenableFile},
		{"detectify", syncDetectifyData, detectifyFile},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wt, fs := newTestWorktree(t)

			commit, err := tc.sync([]byte("1.2.3.4\n"), wt, fs)
			require.NoError(t, err)
			require.NotEqual(t, plumbing.ZeroHash, commit)
			require.Equal(t, "1.2.3.4\n", readFile(t, fs, tc.file))
		})
	}
}
