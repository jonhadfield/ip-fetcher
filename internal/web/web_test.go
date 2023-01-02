package web

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestPathDetails(t *testing.T) {
	// Test a missing file with an missing 'would-be' parent
	output, err := pathDetails(filepath.Join("a", "missing", "directory"))
	require.NoError(t, err)
	require.Equal(t, pathDetailsOutput{
		found:       false,
		parentFound: false,
		isDir:       false,
		parent:      "",
		mode:        0,
	}, output)

	// Test a missing file with a valid parent
	tmpDir := t.TempDir()
	f, err := os.Stat(tmpDir)
	require.NoError(t, err)

	parentMode := f.Mode()

	output, err = pathDetails(filepath.Join(tmpDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, pathDetailsOutput{
		found:       false,
		parentFound: true,
		isDir:       false,
		parent:      filepath.Clean(tmpDir),
		mode:        parentMode,
	}, output)

	// Test a valid directory
	tmpDir = t.TempDir()
	f, err = os.Stat(tmpDir)
	require.NoError(t, err)

	parentMode = f.Mode()

	output, err = pathDetails(tmpDir)
	require.NoError(t, err)
	require.Equal(t, pathDetailsOutput{
		found:       true,
		parentFound: true,
		isDir:       true,
		parent:      filepath.Clean(tmpDir),
		mode:        parentMode,
	}, output)

	// Test a valid file
	tmpDir = t.TempDir()

	filePath := filepath.Join(tmpDir, "testFile")

	file, err := os.Create(filePath)
	require.NoError(t, err)

	_, err = file.WriteString("content")
	require.NoError(t, err)

	f, err = os.Stat(filePath)
	require.NoError(t, err)

	fileMode := f.Mode()

	output, err = pathDetails(filePath)
	require.NoError(t, err)
	require.Equal(t, pathDetailsOutput{
		found:       true,
		parentFound: true,
		isDir:       false,
		parent:      filepath.Clean(tmpDir),
		mode:        fileMode,
	}, output)
}

func TestDownloadFileWithMissingDir(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)
	missingDir := "/a/non-existant-dir"
	downloadedFilePath, err := DownloadFile(client, "https://www.example.com/mytextfile.txt", missingDir)
	require.Error(t, err)
	require.Empty(t, downloadedFilePath)
}

func TestDownloadFileWithDirSpecified(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	tmpDir := t.TempDir()
	df, err := DownloadFile(client, "https://www.example.com/mytextfile.txt", tmpDir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, "mytextfile.txt"), df)
}

func TestDownloadFileWithFilePathSpecified(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	tmpDir := t.TempDir()
	df, err := DownloadFile(client, "https://www.example.com/mytextfile.txt", path.Join(tmpDir, "mytextfile.txt"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, "mytextfile.txt"), df)
}

func TestHTTPGet(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)
	rc := retryablehttp.NewClient()
	rc.HTTPClient = client
	gock.InterceptClient(rc.HTTPClient)
	b, headers, status, err := Request(rc, testUrl, http.MethodGet, nil, nil, 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "hello world", string(b))
	require.Equal(t, "World", headers.Get("hello"))
}
