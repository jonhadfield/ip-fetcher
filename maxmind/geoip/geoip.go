package geoip

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/prefix-fetcher/internal/web"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func New() GeoIP {
	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()
	c.HTTPClient = rc
	c.RetryMax = 1

	return GeoIP{
		Client: c,
	}
}

type GeoIP struct {
	Client     *retryablehttp.Client
	LicenseKey string
	Root       string
	Edition    string
	DBFormat   string // mmdb or csv
	Extract    bool
	Replace    bool
}

const (
	DownloadScheme                        = "https"
	DownloadHost                          = "download.maxmind.com"
	DownloadPath                          = "/app/geoip_download"
	NameASN                               = "ASN"
	NameCountry                           = "Country"
	NameCity                              = "City"
	GeoLite2CityBlocksIPv4CSVFileName     = "GeoLite2-City-Blocks-IPv4.csv"
	GeoLite2CityBlocksIPv6CSVFileName     = "GeoLite2-City-Blocks-IPv6.csv"
	GeoLite2CityLocationsEnCSVFileName    = "GeoLite2-City-Locations-en.csv"
	GeoLite2ASNBlocksIPv4CSVFileName      = "GeoLite2-ASN-Blocks-IPv4.csv"
	GeoLite2ASNBlocksIPv6CSVFileName      = "GeoLite2-ASN-Blocks-IPv6.csv"
	GeoLite2CountryBlocksIPv4CSVFileName  = "GeoLite2-Country-Blocks-IPv4.csv"
	GeoLite2CountryBlocksIPv6CSVFileName  = "GeoLite2-Country-Blocks-IPv6.csv"
	GeoLite2CountryLocationsEnCSVFileName = "GeoLite2-Country-Locations-en.csv"
)

func constructDownloadURL(licenseKey, edition, dbName, dbFormat string) string {
	suffix := "zip"
	if strings.EqualFold(dbFormat, "mmdb") {
		suffix = "tar.gz"
	}

	editionID := fmt.Sprintf("%s-%s", edition, dbName)
	if strings.EqualFold(dbFormat, "csv") {
		editionID = fmt.Sprintf("%s-%s-CSV", edition, dbName)
	}

	u := url.URL{
		Scheme:     DownloadScheme,
		Host:       DownloadHost,
		Path:       DownloadPath,
		ForceQuery: false,
	}

	q := u.Query()
	q.Add("edition_id", editionID)
	q.Add("license_key", licenseKey)
	q.Add("suffix", suffix)
	u.RawQuery = q.Encode()

	return u.String()
}

func fileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}

	return fileName
}

func (gc *GeoIP) Validate() error {
	if gc.LicenseKey == "" {
		return errors.New("missing license key")
	}
	if gc.Root == "" {
		return errors.New("missing download path")
	}

	if gc.Edition == "" {
		return errors.New("missing edition")
	}

	if gc.DBFormat == "" {
		return errors.New("missing database format")
	}

	return nil
}

func (gc *GeoIP) FetchFile(dbName string) (filePath string, err error) {
	if err = gc.Validate(); err != nil {
		return
	}

	downloadURL := constructDownloadURL(gc.LicenseKey, gc.Edition, dbName, gc.DBFormat)

	filename, err := web.RequestContentDispositionFileName(gc.Client, downloadURL, []string{gc.LicenseKey})
	if err != nil {
		return
	}

	filePath = filepath.Join(gc.Root, filename)
	// check if zip already exists
	if _, err = os.Stat(filePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return
		}
	} else {
		// zip already exists
		return
	}

	if _, err = web.DownloadFile(gc.Client, downloadURL, filePath); err != nil {
		return
	}

	return
}

type FetchFilesOutput struct {
	ASNVersion        string
	ASNCompressed     string
	ASNIPv4           string
	ASNIPv6           string
	ASNDataPath       string
	CityVersion       string
	CityIPv4          string
	CityIPv6          string
	CityLocations     string
	CityCompressed    string
	CityDataPath      string
	CountryVersion    string
	CountryIPv4       string
	CountryIPv6       string
	CountryLocations  string
	CountryCompressed string
	CountryDataPath   string
}

func UnzipFiles(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	defer func() {
		if err = r.Close(); err != nil {
			panic(err)
		}
	}()

	if err = os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		var rc io.ReadCloser

		rc, err = f.Open()
		if err != nil {
			return err
		}

		defer func() {
			if err = rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}

			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			defer func() {
				if err = f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, f := range r.File {
		err = extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return err
}

func CheckFileExists(filePath string) (exists bool, err error) {
	if _, err = os.Stat(filePath); err != nil {
		return
	}

	return true, nil
}

func ExtractCountry(zipPath, dest string) error {
	// check if files already exist
	// get expected extracted directory location
	ipv4FileExists, err := CheckFileExists(filepath.Join(dest, GeoLite2CountryBlocksIPv4CSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	ipv6FileExists, err := CheckFileExists(filepath.Join(dest, GeoLite2CountryBlocksIPv6CSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	locationsEnFileNameExists, err := CheckFileExists(filepath.Join(dest, GeoLite2CountryLocationsEnCSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if ipv4FileExists && ipv6FileExists && locationsEnFileNameExists {
		logrus.Debugf("GeoLite2 Country Block CSV files already exist")

		return nil
	}

	return UnzipFiles(zipPath, dest)
}

func ExtractASN(zipPath, dest string) error {
	// check if files already exist
	// get expected extracted directory location
	ipv4FileExists, err := CheckFileExists(filepath.Join(dest, GeoLite2ASNBlocksIPv4CSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	ipv6FileExists, err := CheckFileExists(filepath.Join(dest, GeoLite2ASNBlocksIPv6CSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if ipv4FileExists && ipv6FileExists {
		logrus.Debugf("GeoLite2 ASN Block CSV files already exist")

		return nil
	}

	return UnzipFiles(zipPath, dest)
}

func ExtractCity(zipPath, dest string) error {
	// check if files already exist
	// get expected extracted directory location
	ipv4FileExists, err := CheckFileExists(filepath.Join(dest, GeoLite2CityBlocksIPv4CSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	ipv6FileExists, err := CheckFileExists(filepath.Join(dest, GeoLite2CityBlocksIPv6CSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	locationsEnFileNameExists, err := CheckFileExists(filepath.Join(dest, GeoLite2CityLocationsEnCSVFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if ipv4FileExists && ipv6FileExists && locationsEnFileNameExists {
		logrus.Debugf("GeoLite2 City Block CSV files already exist")

		return nil
	}

	return UnzipFiles(zipPath, dest)
}

func getVersionFromZipFileName(in string) (version string, err error) {
	fNameWithoutExt := fileNameWithoutExtension(filepath.Base(in))
	fNameParts := strings.Split(fNameWithoutExt, "_")
	if len(fNameParts) != 2 {
		err = fmt.Errorf("filename should be in format GeoLite2-<Type>-CSV_YYMMDD.zip but presented was '%s'", in)

		return
	}

	return fNameParts[1], nil
}

func (gc *GeoIP) FetchFiles() (output FetchFilesOutput, err error) {
	if err = gc.Validate(); err != nil {
		return
	}

	output.ASNCompressed, err = gc.FetchFile(NameASN)
	if err != nil {
		return
	}

	if output.ASNVersion, err = getVersionFromZipFileName(output.ASNCompressed); err != nil {
		return
	}

	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractASN(output.ASNCompressed, extractPath); err != nil {
			return
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.ASNCompressed))
		output.ASNDataPath = filepath.Join(extractPath, zipMinusExtension)
		output.ASNIPv4 = filepath.Join(output.ASNDataPath, GeoLite2ASNBlocksIPv4CSVFileName)
		output.ASNIPv6 = filepath.Join(output.ASNDataPath, GeoLite2ASNBlocksIPv6CSVFileName)
	}

	output.CountryCompressed, err = gc.FetchFile(NameCountry)
	if err != nil {
		return
	}

	if output.CountryVersion, err = getVersionFromZipFileName(output.CountryCompressed); err != nil {
		return
	}

	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractCountry(output.CountryCompressed, extractPath); err != nil {
			return
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CountryCompressed))
		output.CountryDataPath = filepath.Join(extractPath, zipMinusExtension)
		output.CountryIPv4 = filepath.Join(extractPath, fileNameWithoutExtension(filepath.Base(output.CountryCompressed)), GeoLite2CountryBlocksIPv4CSVFileName)
		output.CountryIPv6 = filepath.Join(extractPath, fileNameWithoutExtension(filepath.Base(output.CountryCompressed)), GeoLite2CountryBlocksIPv6CSVFileName)
		output.CountryLocations = filepath.Join(extractPath, fileNameWithoutExtension(filepath.Base(output.CountryCompressed)), GeoLite2CountryLocationsEnCSVFileName)
	}

	output.CityCompressed, err = gc.FetchFile(NameCity)
	if err != nil {
		return
	}

	if output.CityVersion, err = getVersionFromZipFileName(output.CityCompressed); err != nil {
		return
	}

	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractCity(output.CityCompressed, extractPath); err != nil {
			return
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CityCompressed))
		output.CityDataPath = filepath.Join(extractPath, zipMinusExtension)
		output.CityIPv4 = filepath.Join(extractPath, fileNameWithoutExtension(filepath.Base(output.CityCompressed)), GeoLite2CityBlocksIPv4CSVFileName)
		output.CityIPv6 = filepath.Join(extractPath, fileNameWithoutExtension(filepath.Base(output.CityCompressed)), GeoLite2CityBlocksIPv6CSVFileName)
		output.CityLocations = filepath.Join(extractPath, fileNameWithoutExtension(filepath.Base(output.CityCompressed)), GeoLite2CityLocationsEnCSVFileName)
	}

	return
}
