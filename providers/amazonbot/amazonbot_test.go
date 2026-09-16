package amazonbot_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/amazonbot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockAmazonbot(t *testing.T) *amazonbot.Amazonbot {
	t.Helper()

	for downloadURL, file := range map[string]string{
		amazonbot.AmazonbotURL: "testdata/amazonbot.html",
		amazonbot.SearchBotURL: "testdata/searchbot.html",
		amazonbot.UserURL:      "testdata/user.html",
	} {
		u, err := url.Parse(downloadURL)
		require.NoError(t, err)

		gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
			Get(u.Path).
			Reply(http.StatusOK).
			File(file)
	}

	a := amazonbot.New()
	gock.InterceptClient(a.Client.HTTPClient)

	return &a
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	a := mockAmazonbot(t)

	doc, err := a.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.Amazonbot.IPv4Prefixes, netip.MustParsePrefix("3.81.245.78/32"))
	require.Contains(t, doc.SearchBot.IPv4Prefixes, netip.MustParsePrefix("3.82.67.224/32"))
	require.Contains(t, doc.User.IPv4Prefixes, netip.MustParsePrefix("100.24.86.21/32"))
	require.NotZero(t, doc.Amazonbot.CreationTime)
}

func TestFetchDataPublishesCombinedJSON(t *testing.T) {
	defer gock.Off()

	a := mockAmazonbot(t)

	data, _, status, err := a.FetchData()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, string(data), `"amazonbot"`)
	require.Contains(t, string(data), `"searchbot"`)
	require.Contains(t, string(data), `"user"`)
	require.NotContains(t, string(data), "<html")
}

func TestExtractDocumentMissing(t *testing.T) {
	_, err := amazonbot.ExtractDocument([]byte("<html>no document</html>"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(amazonbot.AmazonbotURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	a := amazonbot.New()
	gock.InterceptClient(a.Client.HTTPClient)

	_, err = a.Fetch()
	require.Error(t, err)
}
