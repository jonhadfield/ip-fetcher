package amazonbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"regexp"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "amazonbot"
	FullName  = "Amazonbot"
	HostType  = "crawlers"
	SourceURL = "https://developer.amazon.com/amazonbot"
	// AmazonbotURL, SearchBotURL and UserURL are HTML pages that embed the
	// prefix documents. There is no raw JSON endpoint.
	AmazonbotURL = "https://developer.amazon.com/amazonbot/ip-addresses/"
	SearchBotURL = "https://developer.amazon.com/amazonbot/searchbot-ip-addresses/"
	UserURL      = "https://developer.amazon.com/amazonbot/live-ip-addresses/"
)

// errNoDocument is returned when a page carries no extractable prefix document.
var errNoDocument = errors.New("failed to find an amazonbot prefix document on the page")

// documentRegexp matches the JSON object Amazon embeds in each IP list page.
var documentRegexp = regexp.MustCompile(`(?s)\{\s*"creationTime"\s*:\s*"[^"]+"\s*,\s*"prefixes"\s*:\s*\[.*?\]\s*\}`)

type Amazonbot struct {
	Client       *retryablehttp.Client
	AmazonbotURL string
	SearchBotURL string
	UserURL      string
	Timeout      time.Duration
}

func New() Amazonbot {
	return Amazonbot{
		AmazonbotURL: AmazonbotURL,
		SearchBotURL: SearchBotURL,
		UserURL:      UserURL,
		Client:       web.NewHTTPClientWithLogger(),
		Timeout:      web.DefaultRequestTimeout,
	}
}

// RawDoc is the combined on-disk representation of the three upstream lists.
type RawDoc struct {
	Amazonbot json.RawMessage `json:"amazonbot"`
	SearchBot json.RawMessage `json:"searchbot"`
	User      json.RawMessage `json:"user"`
}

type List struct {
	CreationTime time.Time      `json:"creationTime" yaml:"creationTime"`
	IPv4Prefixes []netip.Prefix `json:"ipv4Prefixes" yaml:"ipv4Prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6Prefixes" yaml:"ipv6Prefixes"`
}

// Doc combines the published prefix lists for Amazonbot, Amzn-SearchBot and
// Amzn-User.
type Doc struct {
	Amazonbot List `json:"amazonbot" yaml:"amazonbot"`
	SearchBot List `json:"searchbot" yaml:"searchbot"`
	User      List `json:"user"      yaml:"user"`
}

type rawList struct {
	CreationTime string            `json:"creationTime"`
	Prefixes     []json.RawMessage `json:"prefixes"`
}

type rawPrefix struct {
	IPv4Prefix string `json:"ipv4Prefix"`
	IPv6Prefix string `json:"ipv6Prefix"`
	IPPrefix   string `json:"ip_prefix"`
}

// ExtractDocument returns the JSON document embedded in an Amazonbot IP page.
func ExtractDocument(page []byte) ([]byte, error) {
	match := documentRegexp.Find(page)
	if match == nil {
		return nil, errNoDocument
	}

	if !json.Valid(match) {
		return nil, errNoDocument
	}

	return match, nil
}

func (a *Amazonbot) fetchPage(downloadURL string) ([]byte, http.Header, int, error) {
	data, headers, status, err := web.Request(a.Client, downloadURL, http.MethodGet, nil, nil, a.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download amazonbot addresses from %s. http status code: %d", downloadURL, status)
	}

	doc, err := ExtractDocument(data)
	if err != nil {
		return nil, headers, status, err
	}

	return doc, headers, status, nil
}

// FetchData returns the three extracted documents combined. The HTML pages are
// not published: their markup changes with every site build.
func (a *Amazonbot) FetchData() ([]byte, http.Header, int, error) {
	if a.AmazonbotURL == "" {
		a.AmazonbotURL = AmazonbotURL
	}

	if a.SearchBotURL == "" {
		a.SearchBotURL = SearchBotURL
	}

	if a.UserURL == "" {
		a.UserURL = UserURL
	}

	amazonbot, headers, status, err := a.fetchPage(a.AmazonbotURL)
	if err != nil {
		return nil, headers, status, err
	}

	searchbot, _, searchStatus, err := a.fetchPage(a.SearchBotURL)
	if err != nil {
		return nil, headers, searchStatus, err
	}

	user, _, userStatus, err := a.fetchPage(a.UserURL)
	if err != nil {
		return nil, headers, userStatus, err
	}

	combined, err := json.MarshalIndent(RawDoc{
		Amazonbot: amazonbot,
		SearchBot: searchbot,
		User:      user,
	}, "", "  ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (a *Amazonbot) Fetch() (Doc, error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw RawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	amazonbot, err := processList(raw.Amazonbot)
	if err != nil {
		return Doc{}, err
	}

	searchbot, err := processList(raw.SearchBot)
	if err != nil {
		return Doc{}, err
	}

	user, err := processList(raw.User)
	if err != nil {
		return Doc{}, err
	}

	return Doc{Amazonbot: amazonbot, SearchBot: searchbot, User: user}, nil
}

func processList(data []byte) (List, error) {
	if len(data) == 0 {
		return List{}, nil
	}

	var raw rawList
	if err := json.Unmarshal(data, &raw); err != nil {
		return List{}, err
	}

	list := List{}

	if raw.CreationTime != "" {
		creationTime, err := time.Parse(time.RFC3339Nano, raw.CreationTime)
		if err != nil {
			creationTime, err = time.Parse(time.RFC3339, raw.CreationTime)
			if err != nil {
				return List{}, err
			}
		}

		list.CreationTime = creationTime
	}

	for _, entry := range raw.Prefixes {
		var prefix rawPrefix
		if err := json.Unmarshal(entry, &prefix); err != nil {
			return List{}, err
		}

		candidate := firstNonEmpty(prefix.IPv4Prefix, prefix.IPv6Prefix, prefix.IPPrefix)
		parsed, ok := iplist.ToPrefix(candidate)
		if !ok {
			continue
		}

		if parsed.Addr().Is4() {
			list.IPv4Prefixes = append(list.IPv4Prefixes, parsed)

			continue
		}

		list.IPv6Prefixes = append(list.IPv6Prefixes, parsed)
	}

	return list, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
