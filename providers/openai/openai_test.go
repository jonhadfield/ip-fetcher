package openai_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/openai"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockList(t *testing.T, downloadURL, testDataFile string) {
	t.Helper()

	u, err := url.Parse(downloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File(testDataFile)
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	mockList(t, openai.GPTBotDownloadURL, "testdata/gptbot.json")
	mockList(t, openai.SearchBotDownloadURL, "testdata/searchbot.json")
	mockList(t, openai.ChatGPTUserDownloadURL, "testdata/chatgpt-user.json")

	ac := openai.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)

	require.NotEmpty(t, doc.GPTBot.IPv4Prefixes)
	require.Contains(t, doc.GPTBot.IPv4Prefixes, openai.IPv4Entry{netip.MustParsePrefix("132.196.86.0/24")})
	require.NotEmpty(t, doc.GPTBot.IPv6Prefixes)
	require.Contains(t, doc.GPTBot.IPv6Prefixes, openai.IPv6Entry{netip.MustParsePrefix("2a01:111:f403:c111::/64")})
	require.Equal(t, 2025, doc.GPTBot.CreationTime.Year())

	require.NotEmpty(t, doc.SearchBot.IPv4Prefixes)
	require.Contains(t, doc.SearchBot.IPv4Prefixes, openai.IPv4Entry{netip.MustParsePrefix("104.210.140.128/28")})
	require.Empty(t, doc.SearchBot.IPv6Prefixes)

	require.NotEmpty(t, doc.ChatGPTUser.IPv4Prefixes)
	require.Contains(t, doc.ChatGPTUser.IPv4Prefixes, openai.IPv4Entry{netip.MustParsePrefix("13.65.138.112/28")})
	require.Empty(t, doc.ChatGPTUser.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := openai.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestProcessDataInvalidCreationTime(t *testing.T) {
	_, err := openai.ProcessData([]byte(`{"creationTime": "yesterday", "prefixes": [{"ipv4Prefix": "1.2.3.0/24"}]}`))
	require.Error(t, err)
}
