package dshield_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/providers/dshield"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(dshield.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/block.txt")

	d := dshield.New()
	gock.InterceptClient(d.Client.HTTPClient)

	doc, err := d.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Records, 20)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/block.txt")
	require.NoError(t, err)

	doc, err := dshield.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.Records, 20)

	// the start address and netmask columns give the prefix.
	require.Equal(t, netip.MustParsePrefix("205.210.31.0/24"), doc.Records[0].Prefix)
	require.Equal(t, 339, doc.Records[0].Attacks)
	require.Equal(t, "GOOGLE-CLOUD-PLATFORM", doc.Records[0].Name)
	require.Equal(t, "US", doc.Records[0].Country)
}

// the generation time is carried by an "updated:" header comment.
func TestProcessDataUpdated(t *testing.T) {
	data, err := os.ReadFile("testdata/block.txt")
	require.NoError(t, err)

	doc, err := dshield.ProcessData(data)
	require.NoError(t, err)
	require.Equal(t,
		time.Date(2026, time.August, 20, 14, 0, 24, 391841000, time.UTC),
		doc.Updated)
}

// "-", "None" and ">>UNKNOWN<<" are placeholders for absent registration data.
func TestProcessDataOptionalPlaceholders(t *testing.T) {
	doc, err := dshield.ProcessData([]byte(
		"66.132.195.0\t66.132.195.255\t24\t336\t-\t-\t-\n" +
			"205.210.31.0\t205.210.31.255\t24\t339\tGOOGLE\tUS\tNone\n" +
			"45.194.67.0\t45.194.67.255\t24\t327\tAoC\tZA\t>>UNKNOWN<<\n"))
	require.NoError(t, err)
	require.Len(t, doc.Records, 3)

	require.Empty(t, doc.Records[0].Name)
	require.Empty(t, doc.Records[0].Country)
	require.Empty(t, doc.Records[0].Email)

	require.Equal(t, "GOOGLE", doc.Records[1].Name)
	require.Empty(t, doc.Records[1].Email)

	require.Empty(t, doc.Records[2].Email)
}

// host bits in the start address are masked off by the netmask column.
func TestProcessDataMasksHostBits(t *testing.T) {
	doc, err := dshield.ProcessData([]byte("10.1.2.7\t10.1.2.255\t24\t5\tNET\tGB\tabuse@example.com\n"))
	require.NoError(t, err)
	require.Len(t, doc.Records, 1)
	require.Equal(t, netip.MustParsePrefix("10.1.2.0/24"), doc.Records[0].Prefix)
	require.Equal(t, "abuse@example.com", doc.Records[0].Email)
}

// malformed rows are skipped rather than failing the whole list.
func TestProcessDataSkipsMalformedRows(t *testing.T) {
	doc, err := dshield.ProcessData([]byte(
		"too\tfew\tfields\n" +
			"not-an-ip\t1.2.3.255\t24\t5\tNET\tGB\t-\n" +
			"1.2.3.0\t1.2.3.255\tnotanint\t5\tNET\tGB\t-\n" +
			"1.2.3.0\t1.2.3.255\t24\t5\tNET\tGB\t-\n"))
	require.NoError(t, err)
	require.Len(t, doc.Records, 1)
	require.Equal(t, netip.MustParsePrefix("1.2.3.0/24"), doc.Records[0].Prefix)
}

// a row with a non-numeric attack count still yields a usable prefix.
func TestProcessDataMissingAttackCount(t *testing.T) {
	doc, err := dshield.ProcessData([]byte("1.2.3.0\t1.2.3.255\t24\t\tNET\tGB\t-\n"))
	require.NoError(t, err)
	require.Len(t, doc.Records, 1)
	require.Equal(t, 0, doc.Records[0].Attacks)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(dshield.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	d := dshield.New()
	gock.InterceptClient(d.Client.HTTPClient)

	_, err = d.Fetch()
	require.Error(t, err)
}
