package url

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

type Option func(*Client)

type Client struct {
	Debug bool
	// URLs       []url.URL
	HTTPClient *retryablehttp.Client
}

func New(opt ...Option) *Client {
	pflog.SetLogLevel()

	c := new(Client)

	for _, o := range opt {
		o(c)
	}

	if c.HTTPClient == nil {
		c.HTTPClient = web.NewHTTPClient()
	}

	return c
}

func WithHTTPClient(rc *retryablehttp.Client) Option {
	return func(c *Client) {
		c.HTTPClient = rc
	}
}

type HTTPFiles struct {
	Client *retryablehttp.Client
	URLs   []string
	Debug  bool
}

func NewList() HTTPFiles {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return HTTPFiles{
		URLs:   []string{},
		Client: c,
	}
}

type HTTPFile struct {
	Client *retryablehttp.Client
	URL    string
	Debug  bool
}

func (c *Client) FetchPrefixesAsText(requests []Request) ([]string, error) {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	results := make([]URLResponse, len(requests))

	var mu sync.Mutex
	var fetchErrors []string

	var g errgroup.Group

	for i, req := range requests {
		g.Go(func() error {
			response, err := c.get(req.URL, req.Header)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err.Error())
				mu.Unlock()

				logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())

				return nil //nolint:nilerr
			}

			results[i] = response

			return nil
		})
	}

	_ = g.Wait()

	// Filter out zero-value responses (failed fetches)
	var responses []URLResponse
	for _, r := range results {
		if r.Data != nil {
			responses = append(responses, r)
		}
	}

	pum, err := GetPrefixURLMapFromURLResponses(&responses)
	if err != nil {
		return nil, err
	}

	var prefixes []string
	for k := range pum {
		prefixes = append(prefixes, k.String())
	}

	return prefixes, nil
}

func (hf *HTTPFiles) Add(urls []string) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	for _, u := range urls {
		if _, err := url.Parse(u); err != nil {
			logrus.Infof("%s | failed to parse %s", pflog.GetFunctionName(), u)

			continue
		}

		hf.URLs = append(hf.URLs, u)
	}
}

type RawDoc struct {
	SyncToken     string `json:"syncToken"`
	CreationTime  string `json:"creationTime"`
	LastRequested time.Time
	Entries       []json.RawMessage `json:"prefixes"`
}

type (
	Responses         []URLResponse
	PrefixesWithPaths map[netip.Prefix][]string
)

func (c *Client) FetchPrefixes(requests []Request) (map[netip.Prefix][]string, error) {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	results := make([]URLResponse, len(requests))

	var g errgroup.Group

	for i, req := range requests {
		g.Go(func() error {
			response, err := c.get(req.URL, req.Header)
			if err != nil {
				logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())

				return nil //nolint:nilerr
			}

			results[i] = response

			return nil
		})
	}

	_ = g.Wait()

	// Filter out zero-value responses (failed fetches)
	var responses []URLResponse
	for _, r := range results {
		if r.Data != nil {
			responses = append(responses, r)
		}
	}

	prefixes, err := GetPrefixURLMapFromURLResponses(&responses)
	if err != nil {
		return nil, err
	}

	return prefixes, nil
}

func (hf *HTTPFile) FetchPrefixes() ([]netip.Prefix, error) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	urlResponse, err := hf.FetchURL()
	if err != nil {
		return nil, err
	}

	prefixes, err := ReadRawPrefixesFromURLResponse(urlResponse)
	if err != nil {
		return nil, err
	}

	return prefixes, nil
}

func (hf *HTTPFile) FetchURL() (URLResponse, error) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	result, err := FetchURLResponse(hf.Client, hf.URL)
	if err != nil {
		logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())
	}

	return result, err
}

func extractNetFromString(in string) netip.Prefix {
	funcName := pflog.GetFunctionName()

	r := regexp.MustCompile(`^[0-9a-fA-F](\S+)`)

	s := r.FindString(in)

	// ignore empty strings
	if s == "" {
		return netip.Prefix{}
	}

	if !strings.Contains(s, "/") {
		s += "/32"
	}

	p, err := netip.ParsePrefix(s)
	if err != nil {
		logrus.Tracef("%s | failed to parse %s", funcName, s)
	}

	return p
}

// ReadRawPrefixesFromFileData reads the IPs as strings from the given path.
func ReadRawPrefixesFromFileData(data []byte) ([]netip.Prefix, error) {
	text := strings.Split(string(data), "\n")

	// create regex to check for lines without IPs
	var r *regexp.Regexp
	if len(text) > 0 {
		r = regexp.MustCompile(`^\s*#`)
	}

	var (
		ipnets         []netip.Prefix
		invalidCount   int64
		commentedCount int64
	)
	for _, line := range text {
		// exclude comments
		if r.MatchString(line) {
			commentedCount++

			continue
		}

		if o := extractNetFromString(line); o.IsValid() {
			ipnets = append(ipnets, o)

			continue
		}

		invalidCount++
	}

	logrus.Debugf("%s | loaded %d prefixes from %d lines with %d commented and %d invalid", pflog.GetFunctionName(),
		len(ipnets), len(text), commentedCount, invalidCount)

	return ipnets, nil
}

func ReadRawPrefixesFromURLResponse(response URLResponse) ([]netip.Prefix, error) {
	prefixes, err := ReadRawPrefixesFromFileData(response.Data)

	return prefixes, err
}

func GetPrefixURLMapFromURLResponses(responses *[]URLResponse) (map[netip.Prefix][]string, error) {
	funcName := pflog.GetFunctionName()

	if responses == nil || len(*responses) == 0 {
		return nil, fmt.Errorf("%s | no responses", funcName)
	}

	prefixesWithPaths := make(map[netip.Prefix][]string)

	var responseCount int

	for _, response := range *responses {
		ps, err := ReadRawPrefixesFromURLResponse(response)
		if err != nil {
			logrus.Errorf("%s | %s", funcName, err.Error())
		}

		responseCount++

		logrus.Debugf("loaded %d prefixes from response %s", len(ps), response.url)

		var loadCount int64

		for _, prefix := range ps {
			prefixesWithPaths[prefix] = append(prefixesWithPaths[prefix], response.url)
			loadCount++
		}
	}

	logrus.Debugf("%s | loaded %d unique prefixes from %d response(s)", funcName, len(prefixesWithPaths), responseCount)

	return prefixesWithPaths, nil
}

type DataMap map[string][]string

type URLResponse struct {
	url    string
	Data   []byte
	status int
}

func (c *Client) get(url *url.URL, header http.Header) (URLResponse, error) {
	data, _, status, err := web.Request(
		c.HTTPClient,
		url.String(),
		http.MethodGet,
		header,
		nil,
		web.DefaultRequestTimeout,
	)
	if err != nil {
		logrus.Debug(err.Error())
	}

	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return URLResponse{}, fmt.Errorf("failed to get: %s status: %d", url.String(), status)
	}

	return URLResponse{
		url:    url.String(),
		Data:   data,
		status: status,
	}, err
}

func FetchURLResponse(client *retryablehttp.Client, url string) (URLResponse, error) {
	data, _, status, err := web.Request(client, url, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
	if err != nil {
		logrus.Debug(err.Error())
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return URLResponse{}, fmt.Errorf("failed to fetch: %s status: %d", url, status)
	}

	return URLResponse{
		url:    url,
		Data:   data,
		status: status,
	}, err
}

type GetInput struct {
	URL    url.URL
	Header http.Header
}

type Request struct {
	Method string
	URL    *url.URL
	Header http.Header
}

func (c *Client) Get(requests []Request) (*[]URLResponse, error) {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if len(requests) == 0 {
		return nil, errors.New("no URLs to fetch")
	}

	results := make([]URLResponse, len(requests))

	var g errgroup.Group

	for i, req := range requests {
		g.Go(func() error {
			response, err := c.get(req.URL, req.Header)
			if err != nil {
				logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())

				return fmt.Errorf("%w", err)
			}

			results[i] = response

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &results, nil
}

func (hf *HTTPFiles) FetchURLs() ([]URLResponse, error) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if len(hf.URLs) == 0 {
		return nil, errors.New("no URLs to fetch")
	}

	results := make([]URLResponse, len(hf.URLs))

	var g errgroup.Group

	for i, hfURL := range hf.URLs {
		g.Go(func() error {
			result, err := FetchURLResponse(hf.Client, hfURL)
			if err != nil {
				logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())
			}

			results[i] = result

			return nil
		})
	}

	_ = g.Wait()

	return results, nil
}

type RawIPv4Entry struct {
	IPv4Prefix string `json:"ipv4Prefix"`
	Service    string `json:"service"`
	Scope      string `json:"scope"`
}

type RawIPv6Entry struct {
	IPv6Prefix string `json:"ipv6Prefix"`
	Service    string `json:"service"`
	Scope      string `json:"scope"`
}

type IPv4Entry struct {
	IPv4Prefix netip.Prefix `json:"ipv4Prefix"`
	Service    string       `json:"service"`
	Scope      string       `json:"scope"`
}

type IPv6Entry struct {
	IPv6Prefix netip.Prefix `json:"ipv6Prefix"`
	Service    string       `json:"service"`
	Scope      string       `json:"scope"`
}

type Doc struct {
	SyncToken    string
	CreationTime time.Time
	IPv4Prefixes []IPv4Entry
	IPv6Prefixes []IPv6Entry
}
