package publisher

import (
	_ "embed"
	"fmt"
	"github.com/jonhadfield/ip-fetcher/providers/m247"
	"github.com/jonhadfield/ip-fetcher/providers/scaleway"
	"strings"
	"time"

	"github.com/jonhadfield/ip-fetcher/providers/alibaba"
	"github.com/jonhadfield/ip-fetcher/providers/ovh"

	"github.com/jonhadfield/ip-fetcher/providers/hetzner"
	"github.com/jonhadfield/ip-fetcher/providers/zscaler"

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
)

//go:embed README.template
var ReadMeTemplate string

type Provider struct {
	SyncFunc  func(*git.Worktree, billy.Filesystem) (plumbing.Hash, error)
	ShortName string
	File      string
	FullName  string
	HostType  string
	SourceURL string
}

var providers = []Provider{ //nolint:nolintlint,gochecknoglobals
	{syncAlibaba, alibaba.ShortName, alibabaFile, alibaba.FullName, alibaba.HostType, alibaba.SourceURL},
	{syncAWS, aws.ShortName, awsFile, aws.FullName, aws.HostType, aws.SourceURL},
	{syncAzure, azure.ShortName, azureFile, azure.FullName, azure.HostType, azure.InitialURL},
	{
		syncCloudflare,
		cloudflare.ShortName,
		cloudflareFile,
		cloudflare.FullName,
		cloudflare.HostType,
		cloudflare.SourceURL,
	},
	{syncFastly, fastly.ShortName, fastlyFile, fastly.FullName, fastly.HostType, fastly.SourceURL},
	{syncGCP, gcp.ShortName, gcpFile, gcp.FullName, gcp.HostType, gcp.SourceURL},
	{syncGoogle, google.ShortName, googleFile, google.FullName, google.HostType, google.SourceURL},
	{syncGooglebot, googlebot.ShortName, googlebotFile, googlebot.FullName, googlebot.HostType, googlebot.SourceURL},
	{syncGoogleSC, googlesc.ShortName, googlescFile, googlesc.FullName, googlesc.HostType, googlesc.SourceURL},
	{syncGoogleUTF, googleutf.ShortName, googleutfFile, googleutf.FullName, googleutf.HostType, googleutf.SourceURL},
	{syncHetzner, hetzner.ShortName, hetznerFile, hetzner.FullName, hetzner.HostType, hetzner.SourceURL},
	{syncLinode, linode.ShortName, linodeFile, linode.FullName, linode.HostType, linode.SourceURL},
	{syncM247, m247.ShortName, m247File, m247.FullName, m247.HostType, m247.SourceURL},

	{syncOCI, oci.ShortName, ociFile, oci.FullName, oci.HostType, oci.SourceURL},
	{syncOVH, ovh.ShortName, ovhFile, ovh.FullName, ovh.HostType, ovh.SourceURL},
	{syncScaleway, scaleway.ShortName, scalewayFile, scaleway.FullName, scaleway.HostType, scaleway.SourceURL},
	{syncZscaler, zscaler.ShortName, zscalerFile, zscaler.FullName, zscaler.HostType, zscaler.SourceURL},
}

func GenerateReadMeContent(included []string) (string, error) {
	rows := strings.Builder{}

	for _, inc := range included {
		for _, provider := range providers {
			if inc == provider.ShortName {
				rows.WriteString(
					fmt.Sprintf(
						"| [%s](%s)  | %s |  %s | [source](%s) |  \r\n",
						provider.File,
						provider.File,
						provider.FullName,
						provider.HostType,
						provider.SourceURL,
					),
				)
			}
		}
	}

	content := strings.ReplaceAll(ReadMeTemplate, "{{ date }}", time.Now().UTC().Format(time.RFC1123))
	content = strings.ReplaceAll(content, "{{ rows }}", rows.String())

	return content, nil
}

func syncReadMe(included []string, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	readMeContent, err := GenerateReadMeContent(included)
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
