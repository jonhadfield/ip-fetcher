package bgpview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"golang.org/x/sync/errgroup"
)

// DefaultURL is RIPE stat's announced prefixes endpoint, which is the only
// source this package uses. It previously fell back to api.bgpview.io, which
// has been NXDOMAIN since mid-2026: the fallback could not run, and its dead
// DNS lookups showed up in every error message as a second source that had
// been tried.
const DefaultURL = "https://stat.ripe.net/data/announced-prefixes/data.json?resource=AS%s"

// ripeAttempts is how many times a lookup is tried, each attempt with a
// deadline of its own. One deadline shared across attempts meant a slow
// response burned the whole budget and the retries never ran, so a single
// throttled query failed the provider outright.
const ripeAttempts = 3

// ripeRetryWait is multiplied by the attempt number to space the retries out,
// since the failure being retried is usually per-IP throttling. It is a var so
// the tests can shorten it.
var ripeRetryWait = 2 * time.Second //nolint:gochecknoglobals

// maxConcurrentRIPECalls caps concurrent RIPE stat requests across every caller
// of this package. The publisher fans out 6 BGP-backed providers in parallel and
// each provider has multiple ASNs, which trivially produces a dozen+ simultaneous
// RIPE calls and triggers per-IP throttling. A small global budget keeps us under
// the limit.
const maxConcurrentRIPECalls = 2

var ripeSem = make(chan struct{}, maxConcurrentRIPECalls) //nolint:gochecknoglobals

// Response is the shape the RIPE stat data is mapped into. It began as the
// BGPView API response and is kept as-is, since every caller reads Doc rather
// than this.
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

// RIPEStatResponse represents the RIPE stat API response structure.
type RIPEStatResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Data       struct {
		Prefixes []struct {
			Prefix    string `json:"prefix"`
			Timelines []struct {
				StartTime string `json:"starttime"`
				EndTime   string `json:"endtime"`
			} `json:"timelines"`
		} `json:"prefixes"`
		QueryTime string `json:"query_time"`
		Resource  string `json:"resource"`
	} `json:"data"`
}

// Doc represents a document containing IP prefixes.
type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"IPv4Prefixes"`
	IPv6Prefixes []netip.Prefix `json:"IPv6Prefixes"`
}

// fetchFromRIPEStat fetches data for a single ASN, retrying with a fresh
// deadline each time. The semaphore is released between attempts so a retrying
// lookup does not hold up the others.
func fetchFromRIPEStat(client *retryablehttp.Client, asn, url string, timeout time.Duration) (Response, http.Header, int, error) {
	var (
		response Response
		headers  http.Header
		status   int
		err      error
	)

	for attempt := 1; attempt <= ripeAttempts; attempt++ {
		response, headers, status, err = ripeAttempt(client, asn, url, timeout)
		if err == nil {
			return response, headers, status, nil
		}

		// Only a timeout is worth another attempt. The client already retries a
		// throttled or failed response within the attempt, so looping on those
		// would just multiply requests against an API that is rate-limiting us.
		if !isTimeout(err) {
			return Response{}, headers, status, err
		}

		if attempt < ripeAttempts {
			time.Sleep(time.Duration(attempt) * ripeRetryWait)
		}
	}

	return Response{}, headers, status, fmt.Errorf("gave up after %d attempts: %w", ripeAttempts, err)
}

// isTimeout reports whether the request ran out of time rather than failing
// outright. The deadline is per attempt, so a timeout means this attempt was
// slow, not that the lookup cannot succeed.
func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error

	return errors.As(err, &netErr) && netErr.Timeout()
}

// ripeAttempt is a single lookup, holding the shared budget only while it runs.
func ripeAttempt(client *retryablehttp.Client, asn, url string, timeout time.Duration) (Response, http.Header, int, error) {
	ripeSem <- struct{}{}
	defer func() { <-ripeSem }()

	body, headers, status, err := web.Request(client, url, http.MethodGet, nil, nil, timeout)
	if err != nil {
		return Response{}, nil, 0, fmt.Errorf("error fetching ASN %s from RIPE stat: %w", asn, err)
	}

	if status != http.StatusOK {
		return Response{}, headers, status, fmt.Errorf("RIPE stat API returned status %d for ASN %s", status, asn)
	}

	var ripeResponse RIPEStatResponse
	err = json.Unmarshal(body, &ripeResponse)
	if err != nil {
		return Response{}, nil, 0, fmt.Errorf("error unmarshalling RIPE stat response for ASN %s: %w", asn, err)
	}

	// Convert RIPE stat response to BGPView format
	response := Response{
		Status: ripeResponse.Status,
	}

	for _, prefix := range ripeResponse.Data.Prefixes {
		p, parseErr := netip.ParsePrefix(prefix.Prefix)
		if parseErr != nil {
			continue
		}

		if p.Addr().Is4() {
			response.Data.Ipv4Prefixes = append(response.Data.Ipv4Prefixes, struct {
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
			}{
				Prefix: prefix.Prefix,
				IP:     p.Addr().String(),
				Cidr:   p.Bits(),
			})
		} else if p.Addr().Is6() {
			response.Data.Ipv6Prefixes = append(response.Data.Ipv6Prefixes, struct {
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
			}{
				Prefix: prefix.Prefix,
				IP:     p.Addr().String(),
				Cidr:   p.Bits(),
			})
		}
	}

	return response, headers, status, nil
}

// FetchData fetches IP prefixes for multiple ASNs from the RIPE stat API.
//
// RIPE stat is the authoritative RIR-operated source and, since api.bgpview.io
// went NXDOMAIN in mid-2026, the only one available. A lookup that fails is
// retried rather than falling through to a second source, because there is no
// longer a second source to fall through to.
func FetchData(client *retryablehttp.Client, downloadURL string, asns []string, providerName string, timeout time.Duration) ([]byte, http.Header, int, error) { //nolint:gocognit
	var (
		headers http.Header
		status  int
		err     error
	)

	if downloadURL == "" {
		downloadURL = DefaultURL
	}

	if timeout == 0 {
		timeout = web.LongRequestTimeout
	}

	type asnResult struct {
		response Response
		headers  http.Header
		status   int
	}

	results := make([]asnResult, len(asns))

	// Cap concurrent ASN lookups: RIPE stat appears to rate-limit per-IP and serialises
	// some queries, so blasting 4+ ASNs in parallel turns into timeouts even though each
	// one would individually complete in <1s.
	var g errgroup.Group
	g.SetLimit(maxConcurrentRIPECalls)
	for i, asn := range asns {
		g.Go(func() error {
			asnURL := downloadURL
			if !strings.Contains(asnURL, "%s") {
				asnURL = strings.TrimSuffix(asnURL, "/") + "/%s"
			}

			asnURL = fmt.Sprintf(asnURL, asn)

			response, h, s, ripeErr := fetchFromRIPEStat(client, asn, asnURL, timeout)
			if ripeErr != nil {
				return fmt.Errorf("failed to fetch ASN %s from RIPE stat: %w", asn, ripeErr)
			}

			results[i] = asnResult{response: response, headers: h, status: s}

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return nil, nil, 0, err
	}

	// Use headers/status from last result (matches previous sequential behavior)
	if len(results) > 0 {
		last := results[len(results)-1]
		headers = last.headers
		status = last.status
	}

	doc := Doc{}
	for _, r := range results {
		response := r.response
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

// ProcessData unmarshals the data into a Doc.
func ProcessData(data []byte, providerName string) (Doc, error) {
	var doc Doc
	err := json.Unmarshal(data, &doc)
	if err != nil {
		return Doc{}, fmt.Errorf("error unmarshalling %s doc: %w", providerName, err)
	}

	return doc, nil
}
