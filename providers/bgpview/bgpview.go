package bgpview

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const DefaultURL = "https://api.bgpview.io/asn/%s/prefixes"

// Response represents the BGPView API response structure
type Response struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
	Data          struct {
		Ipv4Prefixes []struct {
			Prefix      string `json:"prefix"`
			IP          string `json:"ip"`
			Cidr        int    `json:"cidr"`
			RoaStatus   string `json:"roa_status"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CountryCode string `json:"country_code"`
			Parent      struct {
				Prefix           string `json:"prefix"`
				IP               string `json:"ip"`
				Cidr             int    `json:"cidr"`
				RirName          string `json:"rir_name"`
				AllocationStatus string `json:"allocation_status"`
			} `json:"parent"`
		} `json:"ipv4_prefixes"`
		Ipv6Prefixes []struct {
			Prefix      string `json:"prefix"`
			IP          string `json:"ip"`
			Cidr        int    `json:"cidr"`
			RoaStatus   string `json:"roa_status"`
			Name        any    `json:"name"`
			Description any    `json:"description"`
			CountryCode any    `json:"country_code"`
			Parent      struct {
				Prefix           any    `json:"prefix"`
				IP               any    `json:"ip"`
				Cidr             any    `json:"cidr"`
				RirName          any    `json:"rir_name"`
				AllocationStatus string `json:"allocation_status"`
			} `json:"parent"`
		} `json:"ipv6_prefixes"`
	} `json:"data"`
	Meta struct {
		TimeZone      string `json:"time_zone"`
		APIVersion    int    `json:"api_version"`
		ExecutionTime string `json:"execution_time"`
	} `json:"@meta"`
}

// Doc represents a document containing IP prefixes
type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"IPv4Prefixes"`
	IPv6Prefixes []netip.Prefix `json:"IPv6Prefixes"`
}

// FetchData fetches IP prefixes for multiple ASNs from BGPView API
func FetchData(client *retryablehttp.Client, downloadURL string, asns []string, providerName string, timeout time.Duration) ([]byte, http.Header, int, error) {
	var (
		headers http.Header
		status  int
		err     error
	)

	if downloadURL == "" {
		downloadURL = DefaultURL
	}

	if timeout == 0 {
		timeout = web.DefaultRequestTimeout
	}

	var responses []Response
	for _, asn := range asns {
		url := downloadURL
		if !strings.Contains(url, "%s") {
			url = strings.TrimSuffix(url, "/") + "/%s"
		}

		url = fmt.Sprintf(url, asn)

		var body []byte
		body, headers, status, err = web.Request(client, url, http.MethodGet, nil, nil, timeout)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("error fetching ASN %s: %w", asn, err)
		}

		if status != http.StatusOK {
			return nil, nil, status, fmt.Errorf("error: ASN %s returned status %d", asn, status)
		}

		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("error unmarshalling ASN %s response: %w", asn, err)
		}

		responses = append(responses, response)
	}

	doc := Doc{}
	for _, response := range responses {
		for _, prefix := range response.Data.Ipv4Prefixes {
			var p netip.Prefix
			p, err = netip.ParsePrefix(prefix.Prefix)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("error parsing IPv4 prefix %s: %w", prefix.Prefix, err)
			}

			doc.IPv4Prefixes = append(doc.IPv4Prefixes, p)
		}

		for _, prefix := range response.Data.Ipv6Prefixes {
			var p netip.Prefix
			p, err = netip.ParsePrefix(prefix.Prefix)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("error parsing IPv6 prefix %s: %w", prefix.Prefix, err)
			}

			doc.IPv6Prefixes = append(doc.IPv6Prefixes, p)
		}
	}

	var jRaw json.RawMessage
	jRaw, err = json.Marshal(doc)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error marshalling doc: %w", err)
	}

	return jRaw, headers, status, nil
}

// ProcessData unmarshals the data into a Doc
func ProcessData(data []byte, providerName string) (Doc, error) {
	var doc Doc
	err := json.Unmarshal(data, &doc)
	if err != nil {
		return Doc{}, fmt.Errorf("error unmarshalling %s doc: %w", providerName, err)
	}

	return doc, nil
}
