package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/gitlab"

	"github.com/urfave/cli/v2"
)

func gitlabCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "gitlab",
		helpName:  "GitLab.com webhook prefixes",
		usage:     "GitLab",
		dataFile:  "gitlab.txt",
		linesFile: "gitlab-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_GITLAB",
		mocks: []mockSource{
			{gitlab.DownloadURL, "../../providers/gitlab/testdata/gitlab.md"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := gitlab.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
