package web_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestRequestMethodNotSpecified(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt?cred=secret-value&hello=world"

	rc := NewHTTPClient()

	body, headers, status, err := Request(rc, testUrl, "", nil, []string{}, 1*time.Second)
	require.Error(t, err)
	require.Empty(t, headers)
	require.Len(t, body, 0)
	require.Equal(t, 0, status)
	require.ErrorContains(t, err, "method not specified")
}

func TestRequestFailure(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt?cred=secret-value&hello=world"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, u.RawQuery)

	gock.New(urlBase).
		Get(u.Path).
		Times(5).
		MatchParam("cred", "secret-value").
		MatchParam("hello", "world").
		ReplyError(errors.New("no worky"))

	c := NewHTTPClient()
	c.RetryMax = 1

	gock.InterceptClient(c.HTTPClient)
	body, headers, status, err := Request(c, testUrl, http.MethodGet, nil, []string{}, 10*time.Second)
	require.Error(t, err)
	require.ErrorContains(t, err, "no worky")
	require.Empty(t, headers)
	require.Len(t, body, 0)
	require.Equal(t, 0, status)
	// require.Empty(t, downloadedFilePath)
}

func TestRequestContentDispositionFileName(t *testing.T) {
	testUrl := "https://www.example.com/sample-url?qsparam=qsvalue"
	fileName := "example-filename.zip"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Head(u.Path).
		MatchParams(map[string]string{
			"qsparam": "qsvalue",
		}).
		Reply(http.StatusOK).
		SetHeader("content-disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	gock.New(urlBase).
		Head(u.Path).
		MatchParams(map[string]string{
			"qsparam": "qsvalue",
		}).
		Reply(http.StatusOK).
		SetHeader("content-disposition", "")

	c := NewHTTPClient()
	c.RetryMax = 0

	gock.InterceptClient(c.HTTPClient)

	fn, err := RequestContentDispositionFileName(c, testUrl, nil)
	require.NoError(t, err)
	require.Equal(t, fileName, fn)

	_, err = RequestContentDispositionFileName(c, testUrl, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get Content-Disposition header")
	require.ErrorContains(t, err, "mime: no media type")

	// second attempt fails as gock only matching one time
	_, err = RequestContentDispositionFileName(c, testUrl, nil)
	require.Error(t, err)
}

func TestPathDetails(t *testing.T) {
	// Test a missing file with an missing 'would-be' parent
	output, err := getPathInfo(filepath.Join("a", "missing", "directory"))
	require.NoError(t, err)
	require.Equal(t, pathInfo{
		exists:       false,
		parentExists: false,
		isDir:        false,
		parent:       "",
		mode:         0,
	}, output)

	// Test a missing file with a valid parent
	tmpDir := t.TempDir()
	f, err := os.Stat(tmpDir)
	require.NoError(t, err)

	parentMode := f.Mode()

	output, err = getPathInfo(filepath.Join(tmpDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, pathInfo{
		exists:       false,
		parentExists: true,
		isDir:        false,
		parent:       filepath.Clean(tmpDir),
		mode:         parentMode,
	}, output)

	// Test a valid directory
	tmpDir = t.TempDir()
	f, err = os.Stat(tmpDir)
	require.NoError(t, err)

	parentMode = f.Mode()

	output, err = getPathInfo(tmpDir)
	require.NoError(t, err)
	require.Equal(t, pathInfo{
		exists:       true,
		parentExists: true,
		isDir:        true,
		parent:       filepath.Clean(tmpDir),
		mode:         parentMode,
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

	output, err = getPathInfo(filePath)
	require.NoError(t, err)
	require.Equal(t, pathInfo{
		exists:       true,
		parentExists: true,
		isDir:        false,
		parent:       filepath.Clean(tmpDir),
		mode:         fileMode,
	}, output)
}

func TestDownloadFileWithMissingDir(t *testing.T) {
	testUrl := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testUrl)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	c := NewHTTPClient()
	gock.InterceptClient(c.HTTPClient)
	c.RetryMax = 0

	missingDir := "/a/non-existant-dir"
	downloadedFilePath, err := DownloadFile(c, "https://www.example.com/mytextfile.txt", missingDir)
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
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	rc := NewHTTPClient()

	gock.InterceptClient(rc.HTTPClient)

	tmpDir := t.TempDir()
	df, err := DownloadFile(rc, "https://www.example.com/mytextfile.txt", tmpDir)
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
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	c := NewHTTPClient()

	gock.InterceptClient(c.HTTPClient)

	tmpDir := t.TempDir()
	df, err := DownloadFile(c, "https://www.example.com/mytextfile.txt", path.Join(tmpDir, "mytextfile.txt"))
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
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	c := NewHTTPClient()

	gock.InterceptClient(c.HTTPClient)

	b, headers, status, err := Request(c, testUrl, http.MethodGet, nil, nil, 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "hello world", string(b))
	require.Equal(t, "World", headers.Get("hello"))
}
