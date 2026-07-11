package openai

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName                = "openai"
	FullName                 = "OpenAI Bots"
	HostType                 = "crawlers"
	SourceURL                = "https://platform.openai.com/docs/bots"
	GPTBotDownloadURL        = "https://openai.com/gptbot.json"
	SearchBotDownloadURL     = "https://openai.com/searchbot.json"
	ChatGPTUserDownloadURL   = "https://openai.com/chatgpt-user.json"
	GPTBotName               = "GPTBot"
	SearchBotName            = "OAI-SearchBot"
	ChatGPTUserName          = "ChatGPT-User"
	downloadedFileTimeFormat = "2006-01-02T15:04:05.999999"
)

func New() OpenAI {
	return OpenAI{
		GPTBotURL:      GPTBotDownloadURL,
		SearchBotURL:   SearchBotDownloadURL,
		ChatGPTUserURL: ChatGPTUserDownloadURL,
		Client:         web.NewHTTPClientWithLogger(),
		Timeout:        web.DefaultRequestTimeout,
	}
}

type OpenAI struct {
	Client         *retryablehttp.Client
	GPTBotURL      string
	SearchBotURL   string
	ChatGPTUserURL string
	Timeout        time.Duration
}

type RawDoc struct {
	CreationTime  string            `json:"creationTime"`
	LastRequested time.Time         `json:"-" yaml:"-"`
	Entries       []json.RawMessage `json:"prefixes"`
}

func (o *OpenAI) FetchGPTBotData() ([]byte, http.Header, int, error) {
	if o.GPTBotURL == "" {
		o.GPTBotURL = GPTBotDownloadURL
	}

	return o.fetchList(o.GPTBotURL)
}

func (o *OpenAI) FetchSearchBotData() ([]byte, http.Header, int, error) {
	if o.SearchBotURL == "" {
		o.SearchBotURL = SearchBotDownloadURL
	}

	return o.fetchList(o.SearchBotURL)
}

func (o *OpenAI) FetchChatGPTUserData() ([]byte, http.Header, int, error) {
	if o.ChatGPTUserURL == "" {
		o.ChatGPTUserURL = ChatGPTUserDownloadURL
	}

	return o.fetchList(o.ChatGPTUserURL)
}

func (o *OpenAI) fetchList(downloadURL string) ([]byte, http.Header, int, error) {
	return web.Request(
		o.Client,
		downloadURL,
		http.MethodGet,
		nil,
		nil,
		o.Timeout,
	)
}

func (o *OpenAI) Fetch() (Doc, error) {
	var doc Doc

	lists := []struct {
		fetch  func() ([]byte, http.Header, int, error)
		target *List
	}{
		{o.FetchGPTBotData, &doc.GPTBot},
		{o.FetchSearchBotData, &doc.SearchBot},
		{o.FetchChatGPTUserData, &doc.ChatGPTUser},
	}

	for _, l := range lists {
		data, _, _, err := l.fetch()
		if err != nil {
			return Doc{}, err
		}

		list, err := ProcessData(data)
		if err != nil {
			return Doc{}, err
		}

		*l.target = list
	}

	return doc, nil
}

func ProcessData(data []byte) (List, error) {
	var (
		list   List
		rawDoc RawDoc
	)

	err := json.Unmarshal(data, &rawDoc)
	if err != nil {
		return List{}, err
	}

	list.IPv4Prefixes, list.IPv6Prefixes, err = castEntries(rawDoc.Entries)
	if err != nil {
		return List{}, err
	}

	ct, err := time.Parse(downloadedFileTimeFormat, rawDoc.CreationTime)
	if err != nil {
		return List{}, err
	}

	list.CreationTime = ct

	return list, nil
}

func castEntries(prefixes []json.RawMessage) ([]IPv4Entry, []IPv6Entry, error) {
	var (
		ipv4 []IPv4Entry
		ipv6 []IPv6Entry
	)

	for _, pr := range prefixes {
		var ipv4entry RawIPv4Entry

		var ipv6entry RawIPv6Entry

		// try 4
		if err := json.Unmarshal(pr, &ipv4entry); err == nil {
			ipv4Prefix, parseError := netip.ParsePrefix(ipv4entry.IPv4Prefix)
			if parseError == nil {
				ipv4 = append(ipv4, IPv4Entry{
					IPv4Prefix: ipv4Prefix,
				})

				continue
			}
		}

		// try 6
		ipv6Err := json.Unmarshal(pr, &ipv6entry)
		if ipv6Err == nil {
			ipv6Prefix, parseError := netip.ParsePrefix(ipv6entry.IPv6Prefix)
			if parseError != nil {
				return ipv4, ipv6, parseError
			}

			ipv6 = append(ipv6, IPv6Entry{
				IPv6Prefix: ipv6Prefix,
			})

			continue
		}

		return ipv4, ipv6, ipv6Err
	}

	return ipv4, ipv6, nil
}

type RawIPv4Entry struct {
	IPv4Prefix string `json:"ipv4Prefix"`
}

type RawIPv6Entry struct {
	IPv6Prefix string `json:"ipv6Prefix"`
}

type IPv4Entry struct {
	IPv4Prefix netip.Prefix `json:"ipv4Prefix"`
}

type IPv6Entry struct {
	IPv6Prefix netip.Prefix `json:"ipv6Prefix"`
}

// List holds the prefixes published for a single OpenAI bot.
type List struct {
	CreationTime time.Time   `json:"creationTime" yaml:"creationTime"`
	IPv4Prefixes []IPv4Entry `json:"ipv4Prefixes" yaml:"ipv4Prefixes"`
	IPv6Prefixes []IPv6Entry `json:"ipv6Prefixes" yaml:"ipv6Prefixes"`
}

// Doc combines the published prefix lists for all OpenAI bots.
type Doc struct {
	GPTBot      List `json:"gptbot" yaml:"gptbot"`
	SearchBot   List `json:"searchbot" yaml:"searchbot"`
	ChatGPTUser List `json:"chatgptUser" yaml:"chatgptUser"`
}
