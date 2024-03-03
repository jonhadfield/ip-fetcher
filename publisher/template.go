package publisher

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/aws"
	"github.com/jonhadfield/ip-fetcher/providers/azure"
	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
	"github.com/jonhadfield/ip-fetcher/providers/fastly"
	"github.com/jonhadfield/ip-fetcher/providers/gcp"
	"github.com/jonhadfield/ip-fetcher/providers/google"
	"github.com/jonhadfield/ip-fetcher/providers/googlebot"
	"github.com/jonhadfield/ip-fetcher/providers/googlesc"
	"github.com/jonhadfield/ip-fetcher/providers/googleutf"
	"github.com/jonhadfield/ip-fetcher/providers/linode"
	"github.com/jonhadfield/ip-fetcher/providers/oci"
	"os"
	"path"
	"strings"
	"time"
)

type Provider struct {
	ShortName string
	File      string
	FullName  string
	HostType  string
	SourceURL string
}

var providers = []Provider{
	{aws.ShortName, awsFile, aws.FullName, aws.HostType, aws.SourceURL},
	{azure.ShortName, azureFile, azure.FullName, azure.HostType, azure.InitialURL},
	{cloudflare.ShortName, cloudflareFile, cloudflare.FullName, cloudflare.HostType, cloudflare.SourceURL},
	{fastly.ShortName, fastlyFile, fastly.FullName, fastly.HostType, fastly.SourceURL},
	{gcp.ShortName, gcpFile, gcp.FullName, gcp.HostType, gcp.SourceURL},
	{google.ShortName, googleFile, google.FullName, google.HostType, google.SourceURL},
	{googlebot.ShortName, googlebotFile, googlebot.FullName, googlebot.HostType, googlebot.SourceURL},
	{googlesc.ShortName, googlescFile, googlesc.FullName, googlesc.HostType, googlesc.SourceURL},
	{googleutf.ShortName, googleutfFile, googleutf.FullName, googleutf.HostType, googleutf.SourceURL},
	{linode.ShortName, linodeFile, linode.FullName, linode.HostType, linode.SourceURL},
	{oci.ShortName, ociFile, oci.FullName, oci.HostType, oci.SourceURL},
}

func generateReadMeContent(included []string) (string, error) {
	templateFile, err := os.ReadFile(path.Join("publisher", "README.template"))
	if err != nil {
		return "", fmt.Errorf("failed to read README template: %w", err)
	}

	rows := strings.Builder{}

	for _, inc := range included {
		for _, provider := range providers {
			if inc == provider.ShortName {
				rows.WriteString(fmt.Sprintf("| [%s](%s)  | %s |  %s | [source](%s) |  \r\n", provider.File, provider.File, provider.FullName, provider.HostType, provider.SourceURL))
			}
		}
	}

	content := strings.ReplaceAll(string(templateFile), "{{ date }}", time.Now().UTC().Format(time.RFC1123))
	content = strings.ReplaceAll(content, "{{ rows }}", rows.String())

	return content, nil
}

func syncReadMe(included []string, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	readMeContent, err := generateReadMeContent(included)
	if err != nil {
		return plumbing.Hash{}, err
	}

	if err = createFile(fs, "README.md", []byte(readMeContent)); err != nil {
		return plumbing.ZeroHash, err
	}

	// Adds the new file to the staging area.
	_, err = wt.Add("README.md")
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update README.md")
}
