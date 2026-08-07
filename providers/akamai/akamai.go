package akamai

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "akamai"
	FullName  = "Akamai"
	HostType  = "cdn"
	SourceURL = "https://techdocs.akamai.com/property-manager/docs/origin-ip-access-control"
	// DownloadURL is the CIDR list Akamai publishes for origin allowlisting:
	// a zip containing akamai_ipv4_CIDRs.txt and akamai_ipv6_CIDRs.txt.
	DownloadURL = "https://techdocs.akamai.com/property-manager/pdfs/akamai_ipv4_ipv6_CIDRs-txt.zip"
)

type Akamai struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Akamai {
	return Akamai{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

func (a *Akamai) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, a.Timeout)
}

func (a *Akamai) Fetch() ([]netip.Prefix, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return nil, err
	}

	return ProcessData(data)
}

// ProcessData parses Akamai CIDR data: either the published zip of
// akamai_ipv4_CIDRs.txt/akamai_ipv6_CIDRs.txt, or a plain text list of
// prefixes, one per line.
func ProcessData(data []byte) ([]netip.Prefix, error) {
	if bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		return processZip(data)
	}

	return parsePrefixLines(data)
}

func processZip(data []byte) ([]netip.Prefix, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read akamai zip: %w", err)
	}

	var prefixes []netip.Prefix

	for _, f := range zr.File {
		// the published zip includes macOS resource-fork entries
		// (__MACOSX/._*.txt) that are not CIDR data
		if !strings.HasSuffix(f.Name, ".txt") ||
			strings.HasPrefix(f.Name, "__MACOSX/") ||
			strings.HasPrefix(path.Base(f.Name), "._") {
			continue
		}

		rc, oErr := f.Open()
		if oErr != nil {
			return nil, fmt.Errorf("failed to open %s in akamai zip: %w", f.Name, oErr)
		}

		content, rErr := io.ReadAll(rc)

		_ = rc.Close()

		if rErr != nil {
			return nil, fmt.Errorf("failed to read %s in akamai zip: %w", f.Name, rErr)
		}

		filePrefixes, pErr := parsePrefixLines(content)
		if pErr != nil {
			return nil, pErr
		}

		prefixes = append(prefixes, filePrefixes...)
	}

	return prefixes, nil
}

func parsePrefixLines(data []byte) ([]netip.Prefix, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var prefixes []netip.Prefix

	for scanner.Scan() {
		// published files carry trailing whitespace on each line
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, err
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}
