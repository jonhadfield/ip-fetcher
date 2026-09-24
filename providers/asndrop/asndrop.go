// Package asndrop retrieves Spamhaus ASN-DROP, the list of ASNs controlled by
// criminal enterprises or known to be hijacked.
package asndrop

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "asndrop"
	FullName  = "Spamhaus ASN-DROP"
	HostType  = "threat"
	SourceURL = "https://www.spamhaus.org/blocklists/do-not-route-or-peer/"
	// DownloadURL returns ASN-DROP as newline delimited JSON objects, ending
	// with a metadata record.
	DownloadURL = "https://www.spamhaus.org/drop/asndrop.json"

	metadataRecordType = "metadata"
)

type ASNDrop struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() ASNDrop {
	return ASNDrop{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawRecord struct {
	Type      string `json:"type,omitempty"`
	ASN       int    `json:"asn,omitempty"`
	RIR       string `json:"rir,omitempty"`
	Domain    string `json:"domain,omitempty"`
	CC        string `json:"cc,omitempty"`
	ASName    string `json:"asname,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Records   int    `json:"records,omitempty"`
}

type Record struct {
	ASN    int    `json:"asn"    yaml:"asn"`
	RIR    string `json:"rir"    yaml:"rir"`
	Domain string `json:"domain" yaml:"domain"`
	CC     string `json:"cc"     yaml:"cc"`
	ASName string `json:"asname" yaml:"asname"`
}

type Doc struct {
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	Records   []Record  `json:"records"   yaml:"records"`
}

func (a *ASNDrop) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, a.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download spamhaus asn-drop from %s. http status code: %d", a.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (a *ASNDrop) Fetch() (Doc, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var doc Doc

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var raw RawRecord
		if err := json.Unmarshal(line, &raw); err != nil {
			return Doc{}, err
		}

		if raw.Type == metadataRecordType {
			if raw.Timestamp != 0 {
				doc.Timestamp = time.Unix(raw.Timestamp, 0).UTC()
			}

			continue
		}

		if raw.ASN == 0 {
			continue
		}

		doc.Records = append(doc.Records, Record{
			ASN:    raw.ASN,
			RIR:    raw.RIR,
			Domain: raw.Domain,
			CC:     raw.CC,
			ASName: raw.ASName,
		})
	}

	if err := scanner.Err(); err != nil {
		return Doc{}, err
	}

	return doc, nil
}

// Lines renders the ASN list as newline separated AS numbers, for --lines.
func (d Doc) Lines() []byte {
	var buf bytes.Buffer
	for _, record := range d.Records {
		fmt.Fprintf(&buf, "AS%d\n", record.ASN)
	}

	return buf.Bytes()
}
