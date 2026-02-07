package geoip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var configureLogrusOnce sync.Once

func configureLogrus() {
	configureLogrusOnce.Do(func() {
		lvl, ok := os.LookupEnv("PF_LOG")
		if !ok {
			lvl = "info"
		}

		ll, err := logrus.ParseLevel(lvl)
		if err != nil {
			ll = logrus.InfoLevel
		}

		logrus.SetLevel(ll)
	})
}

func New() GeoIP {
	configureLogrus()
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

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
	expectedFileNameParts                 = 2
)

const (
	maxFileUncompressedSize    uint64 = 512 << 20 // 512 MiB
	maxArchiveUncompressedSize        = 2 << 30   // 2 GiB
)

func ConstructDownloadURL(licenseKey, edition, dbName, dbFormat string) string {
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

func (gc *GeoIP) FetchFileName(dbName string) (string, error) {
	if err := gc.Validate(); err != nil {
		return "", err
	}

	downloadURL := ConstructDownloadURL(gc.LicenseKey, gc.Edition, dbName, gc.DBFormat)

	return web.RequestContentDispositionFileName(gc.Client, downloadURL, []string{gc.LicenseKey})
}

func (gc *GeoIP) FetchFile(dbName string) (string, error) {
	logrus.Debugf("%s | fetching File %s", pflog.GetFunctionName(), dbName)

	if err := gc.Validate(); err != nil {
		return "", err
	}

	downloadURL := ConstructDownloadURL(gc.LicenseKey, gc.Edition, dbName, gc.DBFormat)

	filename, err := web.RequestContentDispositionFileName(gc.Client, downloadURL, []string{gc.LicenseKey})
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(gc.Root, filename)
	// check if zip already exists
	if _, err = os.Stat(filePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		logrus.Debugf("%s | filepath: %s doesn't exist", pflog.GetFunctionName(), filePath)
	} else {
		logrus.Debugf("%s | filepath: %s already exists", pflog.GetFunctionName(), filePath)
		// zip already exists
		return filePath, nil
	}

	if _, err = web.DownloadFile(gc.Client, downloadURL, filePath); err != nil {
		return "", err
	}

	return filePath, nil
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

func UnzipFiles(src, dest string) (err error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := r.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	if err = os.MkdirAll(dest, 0o750); err != nil {
		return err
	}

	cleanDest := filepath.Clean(dest)
	var totalUncompressed uint64

	for _, file := range r.File {
		extractErr := extractFile(file, cleanDest, &totalUncompressed)
		if extractErr != nil {
			return extractErr
		}
	}

	return nil
}

func extractFile(f *zip.File, dest string, totalUncompressed *uint64) error {
	if f.UncompressedSize64 > maxFileUncompressedSize {
		return fmt.Errorf("file too large: %s", f.Name)
	}

	if *totalUncompressed+f.UncompressedSize64 > maxArchiveUncompressedSize {
		return fmt.Errorf("archive too large: %s", f.Name)
	}

	relativePath := filepath.Clean(f.Name)
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", f.Name)
	}

	targetPath := filepath.Join(dest, relativePath)
	if !strings.HasPrefix(targetPath, dest+string(os.PathSeparator)) && targetPath != dest {
		return fmt.Errorf("illegal file path: %s", f.Name)
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = rc.Close()
	}()

	if f.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, 0o750)
	}

	dirErr := os.MkdirAll(filepath.Dir(targetPath), 0o750)
	if dirErr != nil {
		return dirErr
	}

	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	size := f.UncompressedSize64
	if size > uint64(math.MaxInt64) {
		return fmt.Errorf("file too large: %s", f.Name)
	}

	limitedReader := io.LimitReader(rc, int64(size))
	if _, err = io.Copy(file, limitedReader); err != nil {
		return err
	}

	*totalUncompressed += f.UncompressedSize64

	return nil
}

func CheckFileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		return false, err
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

func GetVersionFromZipFilePath(in string) (string, error) {
	fNameWithoutExt := fileNameWithoutExtension(filepath.Base(in))

	fNameParts := strings.Split(fNameWithoutExt, "_")
	if len(fNameParts) != expectedFileNameParts {
		return "", fmt.Errorf("filename should be in format GeoLite2-<Type>-CSV_YYMMDD.zip but presented was '%s'", in)
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

func (gc *GeoIP) FetchASNFiles() (FetchASNFilesOutput, error) {
	logrus.Debugf("%s | fetching ASN Files", pflog.GetFunctionName())

	var output FetchASNFilesOutput
	var err error
	output.CompressedPath, err = gc.FetchFile(NameASN)
	if err != nil {
		return FetchASNFilesOutput{}, err
	}

	if output.Version, err = GetVersionFromZipFilePath(output.CompressedPath); err != nil {
		return FetchASNFilesOutput{}, err
	}

	if gc.Extract {
		logrus.Debugf("%s | extracting ASN Files", pflog.GetFunctionName())

		extractPath := gc.Root
		if err = ExtractASN(output.CompressedPath, extractPath); err != nil {
			return FetchASNFilesOutput{}, err
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CompressedPath))
		output.DataRoot = filepath.Join(extractPath, zipMinusExtension)
		output.IPv4FilePath = filepath.Join(output.DataRoot, GeoLite2ASNBlocksIPv4CSVFileName)
		output.IPv6FilePath = filepath.Join(output.DataRoot, GeoLite2ASNBlocksIPv6CSVFileName)
	}

	return output, nil
}

type FetchCityFilesOutput struct {
	Version           string
	CompressedPath    string
	DataRoot          string
	IPv4FilePath      string
	IPv6FilePath      string
	LocationsFilePath string
}

func (gc *GeoIP) FetchCityFiles() (FetchCityFilesOutput, error) {
	var output FetchCityFilesOutput
	var err error

	output.CompressedPath, err = gc.FetchFile(NameCity)
	if err != nil {
		return FetchCityFilesOutput{}, err
	}

	if output.Version, err = GetVersionFromZipFilePath(output.CompressedPath); err != nil {
		return FetchCityFilesOutput{}, err
	}
	logrus.Debugf("%s | extracted version %s from %s", pflog.GetFunctionName(), output.Version, output.CompressedPath)
	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractCity(output.CompressedPath, extractPath); err != nil {
			return FetchCityFilesOutput{}, err
		}

		zipMinusExtension := fileNameWithoutExtension(filepath.Base(output.CompressedPath))
		output.DataRoot = filepath.Join(extractPath, zipMinusExtension)
		output.IPv4FilePath = filepath.Join(output.DataRoot, GeoLite2CityBlocksIPv4CSVFileName)
		output.IPv6FilePath = filepath.Join(output.DataRoot, GeoLite2CityBlocksIPv6CSVFileName)
		output.LocationsFilePath = filepath.Join(output.DataRoot, GeoLite2CityLocationsEnCSVFileName)
	}

	return output, nil
}

type FetchCountryFilesOutput struct {
	Version           string
	CompressedPath    string
	DataRoot          string
	IPv4FilePath      string
	IPv6FilePath      string
	LocationsFilePath string
}

func (gc *GeoIP) FetchCountryFiles() (FetchCountryFilesOutput, error) {
	var output FetchCountryFilesOutput
	var err error

	output.CompressedPath, err = gc.FetchFile(NameCountry)
	if err != nil {
		return FetchCountryFilesOutput{}, err
	}

	if output.Version, err = GetVersionFromZipFilePath(output.CompressedPath); err != nil {
		return FetchCountryFilesOutput{}, err
	}

	if gc.Extract {
		extractPath := gc.Root
		if err = ExtractCountry(output.CompressedPath, extractPath); err != nil {
			return FetchCountryFilesOutput{}, err
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

	return output, nil
}

func (gc *GeoIP) FetchAllFiles() (FetchFilesOutput, error) {
	var output FetchFilesOutput
	if err := gc.Validate(); err != nil {
		return FetchFilesOutput{}, err
	}

	var (
		asnOut     FetchASNFilesOutput
		countryOut FetchCountryFilesOutput
		cityOut    FetchCityFilesOutput
	)

	var g errgroup.Group

	g.Go(func() error {
		var err error
		asnOut, err = gc.FetchASNFiles()
		return err
	})
	g.Go(func() error {
		var err error
		countryOut, err = gc.FetchCountryFiles()
		return err
	})
	g.Go(func() error {
		var err error
		cityOut, err = gc.FetchCityFiles()
		return err
	})

	if err := g.Wait(); err != nil {
		return FetchFilesOutput{}, err
	}

	output.ASNCompressedFilePath = asnOut.CompressedPath
	output.ASNIPv4FilePath = asnOut.IPv4FilePath
	output.ASNIPv6FilePath = asnOut.IPv6FilePath
	output.ASNVersion = asnOut.Version

	output.CountryCompressedFilePath = countryOut.CompressedPath
	output.CountryIPv4FilePath = countryOut.IPv4FilePath
	output.CountryIPv6FilePath = countryOut.IPv6FilePath
	output.CountryLocationsFilePath = countryOut.LocationsFilePath
	output.CountryVersion = countryOut.Version

	output.CityCompressedFilePath = cityOut.CompressedPath
	output.CityIPv4FilePath = cityOut.IPv4FilePath
	output.CityIPv6FilePath = cityOut.IPv6FilePath
	output.CityLocationsFilePath = cityOut.LocationsFilePath
	output.CityVersion = cityOut.Version

	return output, nil
}

func (gc *GeoIP) FetchFiles(input FetchFilesInput) (FetchFilesOutput, error) {
	var output FetchFilesOutput
	if err := gc.Validate(); err != nil {
		return FetchFilesOutput{}, err
	}

	var (
		asnOut     FetchASNFilesOutput
		countryOut FetchCountryFilesOutput
		cityOut    FetchCityFilesOutput
	)

	var g errgroup.Group

	if input.ASN {
		g.Go(func() error {
			var err error
			asnOut, err = gc.FetchASNFiles()
			return err
		})
	}

	if input.Country {
		g.Go(func() error {
			var err error
			countryOut, err = gc.FetchCountryFiles()
			return err
		})
	}

	if input.City {
		g.Go(func() error {
			var err error
			cityOut, err = gc.FetchCityFiles()
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return FetchFilesOutput{}, err
	}

	if input.ASN {
		output.ASNDataPath = asnOut.DataRoot
		output.ASNCompressedFilePath = asnOut.CompressedPath
		output.ASNIPv4FilePath = asnOut.IPv4FilePath
		output.ASNIPv6FilePath = asnOut.IPv6FilePath
		output.ASNVersion = asnOut.Version
	}

	if input.Country {
		output.CountryDataPath = countryOut.DataRoot
		output.CountryCompressedFilePath = countryOut.CompressedPath
		output.CountryIPv4FilePath = countryOut.IPv4FilePath
		output.CountryIPv6FilePath = countryOut.IPv6FilePath
		output.CountryLocationsFilePath = countryOut.LocationsFilePath
		output.CountryVersion = countryOut.Version
	}

	if input.City {
		output.CityDataPath = cityOut.DataRoot
		output.CityCompressedFilePath = cityOut.CompressedPath
		output.CityIPv4FilePath = cityOut.IPv4FilePath
		output.CityIPv6FilePath = cityOut.IPv6FilePath
		output.CityLocationsFilePath = cityOut.LocationsFilePath
		output.CityVersion = cityOut.Version
	}

	return output, nil
}
