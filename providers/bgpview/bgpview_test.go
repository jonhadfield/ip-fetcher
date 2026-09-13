// the retry budget is unexported, so these tests live in the package itself.
package bgpview //nolint:testpackage

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const testASN = "60781"

// ripeResponse is the shape RIPE stat returns.
const ripeResponse = `{"status":"ok","data":{"prefixes":[{"prefix":"5.79.0.0/16"},{"prefix":"2a00:c98::/32"}]}}`

func mockRIPE(t *testing.T) string {
	t.Helper()

	u, err := url.Parse(fmt.Sprintf(DefaultURL, testASN))
	require.NoError(t, err)

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func fetchOneASN(t *testing.T) ([]byte, error) {
	t.Helper()

	return fetchWithTimeout(t, time.Second)
}

func fetchWithTimeout(t *testing.T, timeout time.Duration) ([]byte, error) {
	t.Helper()

	client := web.NewHTTPClientWithLogger()
	client.RetryMax = 0
	gock.InterceptClient(client.HTTPClient)

	data, _, _, err := FetchData(client, DefaultURL, []string{testASN}, "test", timeout)

	return data, err
}

// the wait between attempts is real time, and the tests do not need it.
func shortenRetryWait(t *testing.T) {
	t.Helper()

	original := ripeRetryWait
	ripeRetryWait = time.Millisecond

	t.Cleanup(func() { ripeRetryWait = original })
}

func TestFetchData(t *testing.T) {
	defer gock.Off()

	gock.New(mockRIPE(t)).Get("/data/announced-prefixes/data.json").
		Reply(http.StatusOK).BodyString(ripeResponse)

	data, err := fetchOneASN(t)
	require.NoError(t, err)

	doc, err := ProcessData(data, "test")
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("5.79.0.0/16")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2a00:c98::/32")}, doc.IPv6Prefixes)
}

// a slow lookup gets another attempt with a deadline of its own, which is what
// one deadline shared across attempts could not do.
func TestFetchDataRetriesATimeout(t *testing.T) {
	defer gock.Off()

	shortenRetryWait(t)

	gock.New(mockRIPE(t)).Get("/data/announced-prefixes/data.json").
		Reply(http.StatusOK).Delay(200 * time.Millisecond).BodyString(ripeResponse)
	gock.New(mockRIPE(t)).Get("/data/announced-prefixes/data.json").
		Reply(http.StatusOK).BodyString(ripeResponse)

	data, err := fetchWithTimeout(t, 50*time.Millisecond)
	require.NoError(t, err)
	require.Contains(t, string(data), "5.79.0.0/16")
}

// once the attempts are spent the error says so, and names the ASN.
func TestFetchDataGivesUpAfterTheAttempts(t *testing.T) {
	defer gock.Off()

	shortenRetryWait(t)

	gock.New(mockRIPE(t)).Get("/data/announced-prefixes/data.json").
		Times(ripeAttempts).Reply(http.StatusOK).
		Delay(200 * time.Millisecond).BodyString(ripeResponse)

	_, err := fetchWithTimeout(t, 50*time.Millisecond)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("gave up after %d attempts", ripeAttempts))
	require.Contains(t, err.Error(), testASN)
}

// a failure that is not a timeout has already been retried by the client, so it
// is returned rather than looped on.
func TestFetchDataDoesNotRetryAStatusFailure(t *testing.T) {
	defer gock.Off()

	shortenRetryWait(t)

	gock.New(mockRIPE(t)).Get("/data/announced-prefixes/data.json").
		Reply(http.StatusTooManyRequests)

	_, err := fetchOneASN(t)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "gave up after")
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := ProcessData([]byte("not json"), "test")
	require.Error(t, err)
}
