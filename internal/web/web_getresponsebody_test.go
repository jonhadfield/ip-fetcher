package web_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"

	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/stretchr/testify/require"
)

func TestGetResponseBodyPlain(t *testing.T) {
	body := io.NopCloser(bytes.NewBufferString("hello"))
	resp := &http.Response{Body: body, Header: http.Header{}}

	data, err := web.GetResponseBody(resp)
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)
}

func TestGetResponseBodyGzip(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	body := io.NopCloser(&buf)
	resp := &http.Response{Body: body, Header: http.Header{"Content-Encoding": {"gzip"}}}

	data, err := web.GetResponseBody(resp)
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)
}
