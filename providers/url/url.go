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
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

func NewList() HttpFiles {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return HttpFiles{
		Urls:   []string{},
		Client: c,
	}
}

type Option func(*Client)

type Client struct {
	Debug bool
	// URLs       []url.URL
	HttpClient *retryablehttp.Client
}

func New(opt ...Option) *Client {
	pflog.SetLogLevel()

	c := new(Client)

	for _, o := range opt {
		o(c)
	}

	if c.HttpClient == nil {
		c.HttpClient = web.NewHTTPClient()
	}

	return c
}

func WithHttpClient(rc *retryablehttp.Client) Option {
	return func(c *Client) {
		c.HttpClient = rc
	}
}

type HttpFiles struct {
	Client *retryablehttp.Client
	Urls   []string
	Debug  bool
}

type HttpFile struct {
	Client *retryablehttp.Client
	Url    string
	Debug  bool
}

func (c *Client) FetchPrefixesAsText(requests []Request) ([]string, error) {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var (
		responses []UrlResponse
		prefixes  []string
		err       error
	)

	for _, req := range requests {
		var response UrlResponse

		if response, err = c.get(req.Url, req.Header); err != nil {
			logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())

			continue
		}

		responses = append(responses, response)
	}
	pum, err := GetPrefixURLMapFromUrlResponses(&responses)
	if err != nil {
		return nil, err
	}
	for k := range pum {
		prefixes = append(prefixes, k.String())
	}

	return prefixes, nil
}

func (hf *HttpFiles) Add(urls []string) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	for _, u := range urls {
		if _, err := url.Parse(u); err != nil {
			logrus.Infof("%s | failed to parse %s", pflog.GetFunctionName(), u)

			continue
		}

		hf.Urls = append(hf.Urls, u)
	}
}

type RawDoc struct {
	SyncToken     string `json:"syncToken"`
	CreationTime  string `json:"creationTime"`
	LastRequested time.Time
	Entries       []json.RawMessage `json:"prefixes"`
}

type (
	Responses         []UrlResponse
	PrefixesWithPaths map[netip.Prefix][]string
)

func (c *Client) FetchPrefixes(requests []Request) (map[netip.Prefix][]string, error) {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var (
		responses []UrlResponse
		prefixes  map[netip.Prefix][]string
	)

	for _, req := range requests {
		response, err := c.get(req.Url, req.Header)
		if err != nil {
			logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())

			continue
		}

		responses = append(responses, response)
	}

	var err error
	prefixes, err = GetPrefixURLMapFromUrlResponses(&responses)
	if err != nil {
		return nil, err
	}
	//
	// for k := range pum {
	// 	prefixes = append(prefixes, k)
	// }

	return prefixes, nil
}

func (hf *HttpFile) FetchPrefixes() ([]netip.Prefix, error) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	urlResponse, err := hf.FetchUrl()
	if err != nil {
		return nil, err
	}

	prefixes, err := ReadRawPrefixesFromUrlResponse(urlResponse)
	if err != nil {
		return nil, err
	}

	return prefixes, nil
}

func (hf *HttpFile) FetchUrl() (UrlResponse, error) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	result, err := fetchUrlResponse(hf.Client, hf.Url)
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

// ReadRawPrefixesFromFileData reads the IPs as strings from the given path
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

func ReadRawPrefixesFromUrlResponse(response UrlResponse) ([]netip.Prefix, error) {
	prefixes, err := ReadRawPrefixesFromFileData(response.data)

	return prefixes, err
}

func GetPrefixURLMapFromUrlResponses(responses *[]UrlResponse) (map[netip.Prefix][]string, error) {
	funcName := pflog.GetFunctionName()

	if responses == nil || len(*responses) == 0 {
		return nil, fmt.Errorf("%s | no responses", funcName)
	}

	prefixesWithPaths := make(map[netip.Prefix][]string)

	var responseCount int

	for _, response := range *responses {
		ps, err := ReadRawPrefixesFromUrlResponse(response)
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

type UrlResponse struct {
	url    string
	data   []byte
	status int
}

func (c *Client) get(url *url.URL, header http.Header) (UrlResponse, error) {
	data, _, status, err := web.Request(c.HttpClient, url.String(), http.MethodGet, header, nil, web.DefaultRequestTimeout)
	if err != nil {
		logrus.Debug(err.Error())
	}

	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return UrlResponse{}, fmt.Errorf("failed to get: %s status: %d", url.String(), status)
	}

	return UrlResponse{
		url:    url.String(),
		data:   data,
		status: status,
	}, err
}

func fetchUrlResponse(client *retryablehttp.Client, url string) (UrlResponse, error) {
	data, _, status, err := web.Request(client, url, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
	if err != nil {
		logrus.Debug(err.Error())
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return UrlResponse{}, fmt.Errorf("failed to fetch: %s status: %d", url, status)
	}

	return UrlResponse{
		url:    url,
		data:   data,
		status: status,
	}, err
}

type GetInput struct {
	Url    url.URL
	Header http.Header
}

type Request struct {
	Method string
	Url    *url.URL
	Header http.Header
}

func (c *Client) Get(requests []Request) (*[]UrlResponse, error) {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if len(requests) == 0 {
		return nil, errors.New("no URLs to fetch")
	}

	var err error

	var responses []UrlResponse

	for _, req := range requests {
		var response UrlResponse

		if response, err = c.get(req.Url, req.Header); err != nil {
			logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())

			return nil, fmt.Errorf("%w", err)
		}

		responses = append(responses, response)
	}

	return &responses, nil
}

func (hf *HttpFiles) FetchUrls() ([]UrlResponse, error) {
	if hf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if len(hf.Urls) == 0 {
		return nil, errors.New("no URLs to fetch")
	}

	var results []UrlResponse
	for _, hfUrl := range hf.Urls {
		result, err := fetchUrlResponse(hf.Client, hfUrl)
		if err != nil {
			logrus.Debugf("%s | %s", pflog.GetFunctionName(), err.Error())
		}

		results = append(results, result)
	}

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
