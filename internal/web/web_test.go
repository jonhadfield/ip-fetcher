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

	"github.com/jonhadfield/ip-fetcher/internal/web"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestRequestMethodNotSpecified(t *testing.T) {
	testURL := "https://www.example.com/mytextfile.txt?cred=secret-value&hello=world"

	rc := web.NewHTTPClient()

	body, headers, status, err := web.Request(rc, testURL, "", nil, []string{}, 1*time.Second)
	require.Error(t, err)
	require.Empty(t, headers)
	require.Empty(t, body)
	require.Equal(t, 0, status)
	require.ErrorContains(t, err, "method not specified")
}

func TestRequestFailure(t *testing.T) {
	testURL := "https://www.example.com/mytextfile.txt?cred=secret-value&hello=world"
	u, err := url.Parse(testURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, u.RawQuery)

	gock.New(urlBase).
		Get(u.Path).
		Times(5).
		MatchParam("cred", "secret-value").
		MatchParam("hello", "world").
		ReplyError(errors.New("no worky"))

	c := web.NewHTTPClient()
	c.RetryMax = 1

	gock.InterceptClient(c.HTTPClient)
	body, headers, status, err := web.Request(c, testURL, http.MethodGet, nil, []string{}, 10*time.Second)
	require.Error(t, err)
	require.ErrorContains(t, err, "no worky")
	require.Empty(t, headers)
	require.Empty(t, body)
	require.Equal(t, 0, status)
	// require.Empty(t, downloadedFilePath)
}

func TestRequestContentDispositionFileName(t *testing.T) {
	testURL := "https://www.example.com/sample-url?qsparam=qsvalue"
	fileName := "example-filename.zip"
	u, err := url.Parse(testURL)
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

	c := web.NewHTTPClient()
	c.RetryMax = 0

	gock.InterceptClient(c.HTTPClient)

	fn, err := web.RequestContentDispositionFileName(c, testURL, nil)
	require.NoError(t, err)
	require.Equal(t, fileName, fn)

	_, err = web.RequestContentDispositionFileName(c, testURL, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get Content-Disposition header")
	require.ErrorContains(t, err, "mime: no media type")

	// second attempt fails as gock only matching one time
	_, err = web.RequestContentDispositionFileName(c, testURL, nil)
	require.Error(t, err)
}

func TestPathDetails(t *testing.T) {
	// Test a missing file with an missing 'would-be' Parent
	output, err := web.GetPathInfo(filepath.Join("a", "missing", "directory"))
	require.NoError(t, err)
	require.Equal(t, web.PathInfo{
		Exists:       false,
		ParentExists: false,
		IsDir:        false,
		Parent:       "",
		Mode:         0,
	}, output)

	// Test a missing file with a valid Parent
	tmpDir := t.TempDir()
	f, err := os.Stat(tmpDir)
	require.NoError(t, err)

	parentMode := f.Mode()

	output, err = web.GetPathInfo(filepath.Join(tmpDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, web.PathInfo{
		Exists:       false,
		ParentExists: true,
		IsDir:        false,
		Parent:       filepath.Clean(tmpDir),
		Mode:         parentMode,
	}, output)

	// Test a valid directory
	tmpDir = t.TempDir()
	f, err = os.Stat(tmpDir)
	require.NoError(t, err)

	parentMode = f.Mode()

	output, err = web.GetPathInfo(tmpDir)
	require.NoError(t, err)
	require.Equal(t, web.PathInfo{
		Exists:       true,
		ParentExists: true,
		IsDir:        true,
		Parent:       filepath.Clean(tmpDir),
		Mode:         parentMode,
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

	output, err = web.GetPathInfo(filePath)
	require.NoError(t, err)
	require.Equal(t, web.PathInfo{
		Exists:       true,
		ParentExists: true,
		IsDir:        false,
		Parent:       filepath.Clean(tmpDir),
		Mode:         fileMode,
	}, output)
}

func TestDownloadFileWithMissingDir(t *testing.T) {
	testURL := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	c := web.NewHTTPClient()
	gock.InterceptClient(c.HTTPClient)
	c.RetryMax = 0

	missingDir := "/a/non-existant-dir"
	downloadedFilePath, err := web.DownloadFile(c, "https://www.example.com/mytextfile.txt", missingDir)
	require.Error(t, err)
	require.Empty(t, downloadedFilePath)
}

func TestDownloadFileWithDirSpecified(t *testing.T) {
	testURL := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	rc := web.NewHTTPClient()

	gock.InterceptClient(rc.HTTPClient)

	tmpDir := t.TempDir()
	df, err := web.DownloadFile(rc, "https://www.example.com/mytextfile.txt", tmpDir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, "mytextfile.txt"), df)
}

func TestDownloadFileWithFilePathSpecified(t *testing.T) {
	testURL := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	c := web.NewHTTPClient()

	gock.InterceptClient(c.HTTPClient)

	tmpDir := t.TempDir()
	df, err := web.DownloadFile(c, "https://www.example.com/mytextfile.txt", path.Join(tmpDir, "mytextfile.txt"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, "mytextfile.txt"), df)
}

func TestHTTPGet(t *testing.T) {
	testURL := "https://www.example.com/mytextfile.txt"
	u, err := url.Parse(testURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/mytextfile.txt").
		SetHeader("hello", "World")

	c := web.NewHTTPClient()

	gock.InterceptClient(c.HTTPClient)

	b, headers, status, err := web.Request(c, testURL, http.MethodGet, nil, nil, 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "hello world", string(b))
	require.Equal(t, "World", headers.Get("hello"))
}

func TestMaskSecrets(t *testing.T) {
	masked := web.MaskSecrets("token=secret&foo=bar", []string{"secret"})
	require.Equal(t, "token=******&foo=bar", masked)

	masked = web.MaskSecrets("a=1&b=two", nil)
	require.Equal(t, "a=1&b=two", masked)
}
