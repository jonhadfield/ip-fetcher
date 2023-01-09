package geoip

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"gopkg.in/h2non/gock.v1"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstructDownloadURL(t *testing.T) {
	licenseKey := "license-key"
	editionID := "GeoLite2"
	dbType := "ASN"
	csvDBFormat := "cSv"
	mmdbDBFormat := "mmdB"
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?"+
		"edition_id=GeoLite2-ASN-CSV&license_key=license-key&suffix=zip",
		constructDownloadURL(licenseKey, editionID, dbType, csvDBFormat))
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?"+
		"edition_id=GeoLite2-ASN&license_key=license-key&suffix=tar.gz",
		constructDownloadURL(licenseKey, editionID, dbType, mmdbDBFormat))
}

func TestDownloadDBFile(t *testing.T) {
	licenseKey := "test-key"
	tempDir := t.TempDir()

	rc := retryablehttp.NewClient()
	ac := New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "CSv"
	ac.Root = tempDir
	ac.Client = rc
	ac.Client.RetryMax = 0

	downloadURL := constructDownloadURL(licenseKey, "GeoLite2", "ASN", "CSV")
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN-CSV&license_key=test-key&suffix=zip", downloadURL)
	u, err := url.Parse(downloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-ASN-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(u.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-ASN-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-ASN-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(u.Path).
		Reply(200).
		File("testdata/GeoLite2-ASN-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-ASN-CSV_20220617.zip")

	gock.InterceptClient(rc.HTTPClient)

	path, err := ac.FetchFile("ASN")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-ASN-CSV_20220617.zip"), path)
	// require.NotEmpty(t, path.IPv4Prefixes)
}

func TestFetchFiles(t *testing.T) {
	licenseKey := "test-key"
	tempDir := t.TempDir()

	rc := retryablehttp.NewClient()
	ac := New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "csV"
	ac.Extract = true
	ac.Root = tempDir
	ac.Client = rc
	ac.Client.RetryMax = 0

	arnURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN-CSV&license_key=test-key&suffix=zip"
	uARN, _ := url.Parse(arnURL)

	urlBase := fmt.Sprintf("%s://%s", uARN.Scheme, uARN.Host)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-ASN-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(uARN.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-ASN-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-ASN-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(uARN.Path).
		Reply(200).
		File("testdata/GeoLite2-ASN-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-ASN-CSV_20220617.zip")

	cityURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City-CSV&license_key=test-key&suffix=zip"
	uCity, _ := url.Parse(cityURL)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(uCity.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(uCity.Path).
		Reply(200).
		File("testdata/GeoLite2-City-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	countryURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country-CSV&license_key=test-key&suffix=zip"
	uCountry, _ := url.Parse(countryURL)
	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-Country-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(uCountry.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-Country-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-Country-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(uCountry.Path).
		Reply(200).
		File("testdata/GeoLite2-Country-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-Country-CSV_20220617.zip")

	gock.InterceptClient(rc.HTTPClient)

	paths, err := ac.FetchFiles()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-ASN-CSV_20220617.zip"), paths.ASNCompressed)
	require.FileExists(t, paths.ASNCompressed)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-ASN-CSV_20220617", GeoLite2ASNBlocksIPv4CSVFileName), paths.ASNIPv4)
	require.FileExists(t, paths.ASNIPv4)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-ASN-CSV_20220617", GeoLite2ASNBlocksIPv6CSVFileName), paths.ASNIPv6)
	require.FileExists(t, paths.ASNIPv6)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617.zip"), paths.CityCompressed)
	require.FileExists(t, paths.CityCompressed)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617", GeoLite2CityBlocksIPv4CSVFileName), paths.CityIPv4)
	require.FileExists(t, paths.CityIPv4)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617", GeoLite2CityBlocksIPv6CSVFileName), paths.CityIPv6)
	require.FileExists(t, paths.CityIPv6)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617", GeoLite2CityLocationsEnCSVFileName), paths.CityLocations)
	require.FileExists(t, paths.CityLocations)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617.zip"), paths.CountryCompressed)
	require.FileExists(t, paths.CountryCompressed)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617", GeoLite2CountryBlocksIPv4CSVFileName), paths.CountryIPv4)
	require.FileExists(t, paths.CountryIPv4)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617", GeoLite2CountryBlocksIPv6CSVFileName), paths.CountryIPv6)
	require.FileExists(t, paths.CountryIPv6)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617", GeoLite2CountryLocationsEnCSVFileName), paths.CountryLocations)
	require.FileExists(t, paths.CountryLocations)
}

func TestDownloadExtract(t *testing.T) {
	licenseKey := "test-key"
	tempDir := t.TempDir()

	rc := retryablehttp.NewClient()
	ac := New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "csv"
	ac.Root = tempDir
	ac.Client = rc
	ac.Client.RetryMax = 0

	arnURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN-CSV&license_key=test-key&suffix=zip"
	uARN, _ := url.Parse(arnURL)

	urlBase := fmt.Sprintf("%s://%s", uARN.Scheme, uARN.Host)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-ASN-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(uARN.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-ASN-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-ASN-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(uARN.Path).
		Reply(200).
		File("testdata/GeoLite2-ASN-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-ASN-CSV_20220617.zip")

	cityURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City-CSV&license_key=test-key&suffix=zip"
	uCity, _ := url.Parse(cityURL)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(uCity.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(uCity.Path).
		Reply(200).
		File("testdata/GeoLite2-City-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	countryURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country-CSV&license_key=test-key&suffix=zip"
	uCountry, _ := url.Parse(countryURL)
	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-Country-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(uCountry.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-Country-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-Country-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(uCountry.Path).
		Reply(200).
		File("testdata/GeoLite2-Country-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-Country-CSV_20220617.zip")

	gock.InterceptClient(rc.HTTPClient)
	out, err := ac.FetchFiles()
	require.NoError(t, err)

	// require.Equal(t, filepath.Join(tempDir, "GeoLite2-ASN-CSV_20220617.zip"), out.ASN)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617.zip"), out.CityCompressed)
	// require.Equal(t, filepath.Join(tempDir, "GeoLite2-Country-CSV_20220617.zip"), out.Country)
}

// https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN-CSV&license_key=YOUR_LICENSE_KEY&suffix=zip
// "https://download.maxmind.com/app/geoip_download_by_token?edition_id=GeoLite2-ASN-CSV&date=20230106&suffix=zip&token=v2.local.xjWHjBJxD5cDpoXIdshM-xv8z4c_UtzlFWvbFPIiq-06fMCnThhYMVlRC3pWaVRJamUJENteOrKCr7NHoC0rctmqAGKLwbFGLM1jxxD31Hfj97CZ4tGaClG9aXm9OgHzKqsi1pmSjhhIcmVmy_mkEN8XfkNdAcqs6Dw23gGacbvh8fQ5i_bs2SdUkEvXc6ZCxs0Mkw"
// func TestFetch(t *testing.T) {
//	u, err := url.Parse(downloadURL)
//	require.NoError(t, err)
//	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
//
//	gock.New(urlBase).
//		Get(u.Path).
//		Reply(200).
//		// SetHeader("Etag", "cd5e4f079775994d8e49f63ae9a84065").
//		File("testdata/cloud.json")
//
//	rc := retryablehttp.NewClient()
//	ac := New()
//	ac.Client = rc
//	gock.InterceptClient(rc.HTTPClient)
//
//	doc, err := ac.Fetch()
//	require.NoError(t, err)
//	require.NotEmpty(t, doc.IPv4Prefixes)
// }
