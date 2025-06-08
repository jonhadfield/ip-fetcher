package geoip_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/maxmind/geoip"
	"github.com/sirupsen/logrus"
	"gopkg.in/h2non/gock.v1"

	"github.com/stretchr/testify/require"
)

func ensureTestData(t *testing.T, name string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join("testdata", name)); os.IsNotExist(err) {
		t.Skipf("missing %s", name)
	}
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)

	os.Setenv("PF_LOG", "debug")

	// run tests
	code := m.Run()

	os.Exit(code)
}

// TODO: test for DB type other than all caps
func TestConstructDownloadURL(t *testing.T) {
	licenseKey := "license-key"
	editionID := "GeoLite2"
	dbType := "ASN"
	csvDBFormat := "cSv"
	mmdbDBFormat := "mmdB"
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?"+
		"edition_id=GeoLite2-ASN-CSV&license_key=license-key&suffix=zip",
		geoip.ConstructDownloadURL(licenseKey, editionID, dbType, csvDBFormat))
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?"+
		"edition_id=GeoLite2-ASN&license_key=license-key&suffix=tar.gz",
		geoip.ConstructDownloadURL(licenseKey, editionID, dbType, mmdbDBFormat))
}

func TestGetVersionFromZipFileName(t *testing.T) {
	v, err := geoip.GetVersionFromZipFilePath("GeoLite2-Country-CSV_20220617.zip")
	require.NoError(t, err)
	require.Equal(t, "20220617", v)
	v, err = geoip.GetVersionFromZipFilePath("/tmp/some/other/dir/GeoLite2-Country-CSV_20220617.zip")
	require.NoError(t, err)
	require.Equal(t, "20220617", v)
}

func TestDownloadDBFile(t *testing.T) {
	ensureTestData(t, "GeoLite2-ASN-CSV_20220617.zip")
	licenseKey := "test-key"
	tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "CSv"
	ac.Root = tempDir
	ac.Client.RetryMax = 0

	downloadURL := geoip.ConstructDownloadURL(licenseKey, "GeoLite2", "ASN", "CSV")
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

	gock.InterceptClient(ac.Client.HTTPClient)

	path, err := ac.FetchFile("ASN")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-ASN-CSV_20220617.zip"), path)
}

func TestDownloadDBFileMissingTargetDirectory(t *testing.T) {
	ensureTestData(t, "GeoLite2-ASN-CSV_20220617.zip")
	licenseKey := "test-key"
	tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "CSv"
	ac.Root = filepath.Join(tempDir, "invalid")
	ac.Client.RetryMax = 0

	downloadURL := geoip.ConstructDownloadURL(licenseKey, "GeoLite2", "ASN", "CSV")
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

	gock.InterceptClient(ac.Client.HTTPClient)

	_, err = ac.FetchFile("ASN")
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid")
	require.ErrorContains(t, err, "exist")
}

func TestFetchCityFiles(t *testing.T) {
	ensureTestData(t, "GeoLite2-City-CSV_20220617.zip")
	licenseKey := "test-key"
	tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "cSv"
	ac.Extract = true
	ac.Root = tempDir
	ac.Client.RetryMax = 0

	downloadURL := geoip.ConstructDownloadURL(licenseKey, "GeoLite2", "ciTy", "CSV")
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City-CSV&license_key=test-key&suffix=zip", downloadURL)
	u, err := url.Parse(downloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(u.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(u.Path).
		Reply(200).
		File("testdata/GeoLite2-City-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	gock.InterceptClient(ac.Client.HTTPClient)

	o, err := ac.FetchCityFiles()
	require.Equal(t, "20220617", o.Version)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617.zip"), o.CompressedPath)
	require.FileExists(t, o.CompressedPath)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617/"), o.DataRoot)
	require.DirExists(t, o.DataRoot)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617/GeoLite2-City-Blocks-IPv4.csv"), o.IPv4FilePath)
	require.FileExists(t, o.IPv4FilePath)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617/GeoLite2-City-Blocks-IPv6.csv"), o.IPv6FilePath)
	require.FileExists(t, o.IPv6FilePath)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617/GeoLite2-City-Locations-en.csv"), o.LocationsFilePath)
	require.FileExists(t, o.LocationsFilePath)
}

func TestFetchCityFilesWithoutExtract(t *testing.T) {
	ensureTestData(t, "GeoLite2-City-CSV_20220617.zip")
	licenseKey := "test-key"
	tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "cSv"
	ac.Extract = false
	ac.Root = tempDir
	ac.Client.RetryMax = 0

	downloadURL := geoip.ConstructDownloadURL(licenseKey, "GeoLite2", "ciTy", "CSV")
	require.Equal(t, "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City-CSV&license_key=test-key&suffix=zip", downloadURL)
	u, err := url.Parse(downloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Head(u.Path).
		Reply(200).
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	gock.New(urlBase).
		MatchParams(map[string]string{
			"edition_id":  "GeoLite2-City-CSV",
			"license_key": "test-key",
			"suffix":      "zip",
		}).
		Get(u.Path).
		Reply(200).
		File("testdata/GeoLite2-City-CSV_20220617.zip").
		SetHeader("content-disposition", "attachment; filename=GeoLite2-City-CSV_20220617.zip")

	gock.InterceptClient(ac.Client.HTTPClient)

	o, err := ac.FetchCityFiles()
	require.Equal(t, "20220617", o.Version)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617.zip"), o.CompressedPath)
	require.FileExists(t, o.CompressedPath)
	require.Empty(t, o.DataRoot)
	require.NoDirExists(t, o.DataRoot)
	require.Empty(t, o.IPv4FilePath)
	require.NoFileExists(t, o.IPv4FilePath)
	require.Empty(t, o.IPv6FilePath)
	require.NoFileExists(t, o.IPv6FilePath)
	require.Empty(t, o.LocationsFilePath)
	require.NoFileExists(t, o.LocationsFilePath)
}

func TestFetchFiles(t *testing.T) {
	ensureTestData(t, "GeoLite2-ASN-CSV_20220617.zip")
	ensureTestData(t, "GeoLite2-City-CSV_20220617.zip")
	ensureTestData(t, "GeoLite2-Country-CSV_20220617.zip")
	licenseKey := "test-key"
	tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "csV"
	ac.Extract = true
	ac.Root = tempDir
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

	gock.InterceptClient(ac.Client.HTTPClient)

	paths, err := ac.FetchFiles(geoip.FetchFilesInput{
		ASN:     true,
		Country: true,
		City:    true,
	})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-ASN-CSV_20220617.zip"), paths.ASNCompressedFilePath)
	require.FileExists(t, paths.ASNCompressedFilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-ASN-CSV_20220617", geoip.GeoLite2ASNBlocksIPv4CSVFileName), paths.ASNIPv4FilePath)
	require.FileExists(t, paths.ASNIPv4FilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-ASN-CSV_20220617", geoip.GeoLite2ASNBlocksIPv6CSVFileName), paths.ASNIPv6FilePath)
	require.FileExists(t, paths.ASNIPv6FilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617.zip"), paths.CityCompressedFilePath)
	require.FileExists(t, paths.CityCompressedFilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617", geoip.GeoLite2CityBlocksIPv4CSVFileName), paths.CityIPv4FilePath)
	require.FileExists(t, paths.CityIPv4FilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617", geoip.GeoLite2CityBlocksIPv6CSVFileName), paths.CityIPv6FilePath)
	require.FileExists(t, paths.CityIPv6FilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-City-CSV_20220617", geoip.GeoLite2CityLocationsEnCSVFileName), paths.CityLocationsFilePath)
	require.FileExists(t, paths.CityLocationsFilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617.zip"), paths.CountryCompressedFilePath)
	require.FileExists(t, paths.CountryCompressedFilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617", geoip.GeoLite2CountryBlocksIPv4CSVFileName), paths.CountryIPv4FilePath)
	require.FileExists(t, paths.CountryIPv4FilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617", geoip.GeoLite2CountryBlocksIPv6CSVFileName), paths.CountryIPv6FilePath)
	require.FileExists(t, paths.CountryIPv6FilePath)
	require.Equal(t, filepath.Join(ac.Root, "GeoLite2-Country-CSV_20220617", geoip.GeoLite2CountryLocationsEnCSVFileName), paths.CountryLocationsFilePath)
	require.FileExists(t, paths.CountryLocationsFilePath)
}

func TestDownloadExtract(t *testing.T) {
	ensureTestData(t, "GeoLite2-ASN-CSV_20220617.zip")
	ensureTestData(t, "GeoLite2-City-CSV_20220617.zip")
	licenseKey := "test-key"
	tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "csv"
	ac.Root = tempDir
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

	gock.InterceptClient(ac.Client.HTTPClient)
	out, err := ac.FetchFiles(geoip.FetchFilesInput{
		ASN:     true,
		Country: true,
		City:    true,
	})
	require.NoError(t, err)

	require.Equal(t, filepath.Join(tempDir, "GeoLite2-ASN-CSV_20220617.zip"), out.ASNCompressedFilePath)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-City-CSV_20220617.zip"), out.CityCompressedFilePath)
	require.Equal(t, filepath.Join(tempDir, "GeoLite2-Country-CSV_20220617.zip"), out.CountryCompressedFilePath)
}

func TestFetchCityFilesEmptyRoot(t *testing.T) {
	licenseKey := "test-key"
	// tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "cSv"
	ac.Extract = true
	ac.Client.RetryMax = 0

	_, err := ac.FetchCityFiles()
	require.ErrorContains(t, err, "missing download path")
}

func TestDownloadExtractCityWithoutRoot(t *testing.T) {
	ensureTestData(t, "GeoLite2-City-CSV_20220617.zip")
	licenseKey := "test-key"
	// tempDir := t.TempDir()

	ac := geoip.New()
	ac.LicenseKey = licenseKey
	ac.Edition = "GeoLite2"
	ac.DBFormat = "csv"
	ac.Root = "/tmp"
	ac.Client.RetryMax = 0

	arnURL := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN-CSV&license_key=test-key&suffix=zip"
	uARN, _ := url.Parse(arnURL)

	urlBase := fmt.Sprintf("%s://%s", uARN.Scheme, uARN.Host)

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

	gock.InterceptClient(ac.Client.HTTPClient)
	out, err := ac.FetchFiles(geoip.FetchFilesInput{
		ASN:     false,
		Country: false,
		City:    true,
	})
	require.NoError(t, err)

	require.Equal(t, filepath.Join("/tmp", "GeoLite2-City-CSV_20220617.zip"), out.CityCompressedFilePath)
}
