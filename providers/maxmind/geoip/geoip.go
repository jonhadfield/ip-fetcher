package geoip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

func init() {
	lvl, ok := os.LookupEnv("PF_LOG")
	// LOG_LEVEL not set, default to info
	if !ok {
		lvl = "info"
	}

	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.InfoLevel
	}

	logrus.SetLevel(ll)
}

func New() GeoIP {
	pflog.SetLogLevel()
	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}
	c.HTTPClient = rc
	c.RetryMax = 1

	return GeoIP{
		Client: c,
	}
}

type LeveledLogrus struct {
	*logrus.Logger
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

	switch strings.ToLower(dbName) {
	case "country":
		dbName = "Country"
	case "asn":
		dbName = "ASN"
	case "city":
		dbName = "City"
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
	if _, err := os.Stat(gc.Root); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("data root '%s' doesn't exist", gc.Root)
		}

		return fmt.Errorf("failed to access data root '%s' - %w", gc.Root, err)
	}

	if gc.Edition == "" {
		return errors.New("missing edition")
	}

	if gc.DBFormat == "" {
		return errors.New("missing database format")
	}

	return nil
}

func (gc *GeoIP) FetchFileName(dbName string) (filename string, err error) {
	if err = gc.Validate(); err != nil {
		return
	}

	downloadURL := constructDownloadURL(gc.LicenseKey, gc.Edition, dbName, gc.DBFormat)

	return web.RequestContentDispositionFileName(gc.Client, downloadURL, []string{gc.LicenseKey})
}

func (gc *GeoIP) FetchFile(dbName string) (filePath string, err error) {
	logrus.Debugf("%s | fetching File %s", pflog.GetFunctionName(), dbName)

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

		logrus.Debugf("%s | filepath: %s doesn't exist", pflog.GetFunctionName(), filePath)
	} else {
		logrus.Debugf("%s | filepath: %s already exists", pflog.GetFunctionName(), filePath)
		// zip already exists
		return
	}

	if _, err = web.DownloadFile(gc.Client, downloadURL, filePath); err != nil {
		return
	}

	return
}

type FetchFilesOutput struct {
	ASNVersion                string
	ASNCompressedFilePath     string
	ASNIPv4FilePath           string
	ASNIPv6FilePath           string
	ASNDataPath               string
	CityVersion               string
	CityIPv4FilePath          string
	CityIPv6FilePath          string
	CityLocationsFilePath     string
	CityCompressedFilePath    string
	CityDataPath              string
	CountryVersion            string
	CountryIPv4FilePath       string
	CountryIPv6FilePath       string
	CountryLocationsFilePath  string
	CountryCompressedFilePath string
	CountryDataPath           string
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

func getVersionFromZipFilePath(in string) (version string, err error) {
	fNameWithoutExt := fileNameWithoutExtension(filepath.Base(in))

	fNameParts := strings.Split(fNameWithoutExt, "_")
	if len(fNameParts) != 2 {
		err = fmt.Errorf("filename should be in format GeoLite2-<Type>-CSV_YYMMDD.zip but presented was '%s'", in)

		return
	}

	return fNameParts[1], nil
}

type FetchFilesInput struct {
	ASN     bool
	Country bool
	City    bool
}

type FetchASNFilesOutput struct {
	Version        string
	CompressedPath string
	DataRoot       string
	IPv4FilePath   string
	IPv6FilePath   string
}

func (gc *GeoIP) FetchASNFiles() (output FetchASNFilesOutput, err error) {
	logrus.Debugf("%s | fetching ASN Files", pflog.GetFunctionName())
	output.CompressedPath, err = gc.FetchFile(NameASN)
	if err != nil {
		return
	}

	if output.Version, err = getVersionFromZipFilePath(output.CompressedPath); err != nil {
		return
	}

	if gc.Extract {
		logrus.Debugf("%s | extracting ASN Files", pflog.GetFunctionName())

		extractPath := gc.Root
		if err = ExtractASN(output.CompressedPath, extractPath); err != nil {
			return
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CompressedPath))
		output.DataRoot = filepath.Join(extractPath, zipMinusExtension)
		output.IPv4FilePath = filepath.Join(output.DataRoot, GeoLite2ASNBlocksIPv4CSVFileName)
		output.IPv6FilePath = filepath.Join(output.DataRoot, GeoLite2ASNBlocksIPv6CSVFileName)
	}

	return output, err
}

type FetchCityFilesOutput struct {
	Version           string
	CompressedPath    string
	DataRoot          string
	IPv4FilePath      string
	IPv6FilePath      string
	LocationsFilePath string
}

func (gc *GeoIP) FetchCityFiles() (output FetchCityFilesOutput, err error) {
	output.CompressedPath, err = gc.FetchFile(NameCity)
	if err != nil {
		return
	}

	if output.Version, err = getVersionFromZipFilePath(output.CompressedPath); err != nil {
		return
	}
	logrus.Debugf("%s | extracted version %s from %s", pflog.GetFunctionName(), output.Version, output.CompressedPath)
	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractCity(output.CompressedPath, extractPath); err != nil {
			return
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CompressedPath))
		output.DataRoot = filepath.Join(extractPath, zipMinusExtension)
		output.IPv4FilePath = filepath.Join(output.DataRoot, GeoLite2CityBlocksIPv4CSVFileName)
		output.IPv6FilePath = filepath.Join(output.DataRoot, GeoLite2CityBlocksIPv6CSVFileName)
		output.LocationsFilePath = filepath.Join(output.DataRoot, GeoLite2CityLocationsEnCSVFileName)
	}

	return output, err
}

type FetchCountryFilesOutput struct {
	Version           string
	CompressedPath    string
	DataRoot          string
	IPv4FilePath      string
	IPv6FilePath      string
	LocationsFilePath string
}

func (gc *GeoIP) FetchCountryFiles() (output FetchCountryFilesOutput, err error) {
	output.CompressedPath, err = gc.FetchFile(NameCountry)
	if err != nil {
		return
	}

	if output.Version, err = getVersionFromZipFilePath(output.CompressedPath); err != nil {
		return
	}

	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractCountry(output.CompressedPath, extractPath); err != nil {
			return
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CompressedPath))
		output.DataRoot = filepath.Join(extractPath, zipMinusExtension)
		logrus.Debugf("maxmind data root: %s", output.DataRoot)
		output.IPv4FilePath = filepath.Join(output.DataRoot, GeoLite2CountryBlocksIPv4CSVFileName)
		logrus.Debugf("GeoLite2CountryBlocksIPv4CSVFileName path: %s", output.IPv4FilePath)
		output.IPv6FilePath = filepath.Join(output.DataRoot, GeoLite2CountryBlocksIPv6CSVFileName)
		logrus.Debugf("GeoLite2CountryBlocksIPv6CSVFileName path: %s", output.IPv6FilePath)
		output.LocationsFilePath = filepath.Join(output.DataRoot, GeoLite2CountryLocationsEnCSVFileName)
		logrus.Debugf("GeoLite2CountryLocationsEnCSVFileName path: %s", output.LocationsFilePath)
	}

	return output, err
}

func (gc *GeoIP) FetchAllFiles() (output FetchFilesOutput, err error) {
	if err = gc.Validate(); err != nil {
		return
	}

	var asnOut FetchASNFilesOutput
	asnOut, err = gc.FetchASNFiles()
	if err != nil {
		return
	}

	output.ASNCompressedFilePath = asnOut.CompressedPath
	output.ASNIPv4FilePath = asnOut.IPv4FilePath
	output.ASNIPv6FilePath = asnOut.IPv6FilePath
	output.ASNVersion = asnOut.Version

	var CountryOut FetchCountryFilesOutput
	CountryOut, err = gc.FetchCountryFiles()
	if err != nil {
		return
	}

	output.CountryCompressedFilePath = CountryOut.CompressedPath
	output.CountryIPv4FilePath = CountryOut.IPv4FilePath
	output.CountryIPv6FilePath = CountryOut.IPv6FilePath
	output.CountryLocationsFilePath = CountryOut.LocationsFilePath
	output.CountryVersion = CountryOut.Version

	var CityOut FetchCityFilesOutput
	CityOut, err = gc.FetchCityFiles()
	if err != nil {
		logrus.Errorf("%s | %s", pflog.GetFunctionName(), err.Error())

		return
	}

	output.CityCompressedFilePath = CityOut.CompressedPath
	output.CityIPv4FilePath = CityOut.IPv4FilePath
	output.CityIPv6FilePath = CityOut.IPv6FilePath
	output.CityLocationsFilePath = CityOut.LocationsFilePath
	output.CityVersion = CityOut.Version

	return
}

func (gc *GeoIP) FetchFiles(input FetchFilesInput) (output FetchFilesOutput, err error) {
	if err = gc.Validate(); err != nil {
		return
	}

	if input.ASN {
		var asnOut FetchASNFilesOutput
		asnOut, err = gc.FetchASNFiles()
		if err != nil {
			return
		}

		output.ASNDataPath = asnOut.DataRoot
		output.ASNCompressedFilePath = asnOut.CompressedPath
		output.ASNIPv4FilePath = asnOut.IPv4FilePath
		output.ASNIPv6FilePath = asnOut.IPv6FilePath
		output.ASNVersion = asnOut.Version
	}

	if input.Country {
		var CountryOut FetchCountryFilesOutput
		CountryOut, err = gc.FetchCountryFiles()
		if err != nil {
			return
		}

		output.CountryDataPath = CountryOut.DataRoot
		output.CountryCompressedFilePath = CountryOut.CompressedPath
		output.CountryIPv4FilePath = CountryOut.IPv4FilePath
		output.CountryIPv6FilePath = CountryOut.IPv6FilePath
		output.CountryLocationsFilePath = CountryOut.LocationsFilePath
		output.CountryVersion = CountryOut.Version
	}

	if input.City {
		var CityOut FetchCityFilesOutput
		CityOut, err = gc.FetchCityFiles()
		if err != nil {
			return
		}

		output.CityDataPath = CityOut.DataRoot
		output.CityCompressedFilePath = CityOut.CompressedPath
		output.CityIPv4FilePath = CityOut.IPv4FilePath
		output.CityIPv6FilePath = CityOut.IPv6FilePath
		output.CityLocationsFilePath = CityOut.LocationsFilePath
		output.CityVersion = CityOut.Version
	}

	return
}
