package publisher

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/jonhadfield/ip-fetcher/providers/m247"
	"github.com/jonhadfield/ip-fetcher/providers/scaleway"
	"github.com/jonhadfield/ip-fetcher/providers/vultr"

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
	FetchFunc    func() ([]byte, error)
	SyncDataFunc func(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error)
	ShortName    string
	File         string
	FullName     string
	HostType     string
	SourceURL    string
}

var providers = []Provider{ //nolint:nolintlint,gochecknoglobals
	{fetchAlibaba, syncAlibabaData, alibaba.ShortName, alibabaFile, alibaba.FullName, alibaba.HostType, alibaba.SourceURL},
	{fetchAWS, syncAWSData, aws.ShortName, awsFile, aws.FullName, aws.HostType, aws.SourceURL},
	{fetchAzure, syncAzureData, azure.ShortName, azureFile, azure.FullName, azure.HostType, azure.InitialURL},
	{fetchCloudflare, syncCloudflareData, cloudflare.ShortName, cloudflareFile, cloudflare.FullName, cloudflare.HostType, cloudflare.SourceURL},
	{fetchFastly, syncFastlyData, fastly.ShortName, fastlyFile, fastly.FullName, fastly.HostType, fastly.SourceURL},
	{fetchGCP, syncGCPData, gcp.ShortName, gcpFile, gcp.FullName, gcp.HostType, gcp.SourceURL},
	{fetchGoogle, syncGoogleData, google.ShortName, googleFile, google.FullName, google.HostType, google.SourceURL},
	{fetchGooglebot, syncGooglebotData, googlebot.ShortName, googlebotFile, googlebot.FullName, googlebot.HostType, googlebot.SourceURL},
	{fetchGoogleSC, syncGoogleSCData, googlesc.ShortName, googlescFile, googlesc.FullName, googlesc.HostType, googlesc.SourceURL},
	{fetchGoogleUTF, syncGoogleUTFData, googleutf.ShortName, googleutfFile, googleutf.FullName, googleutf.HostType, googleutf.SourceURL},
	{fetchHetzner, syncHetznerData, hetzner.ShortName, hetznerFile, hetzner.FullName, hetzner.HostType, hetzner.SourceURL},
	{fetchLinode, syncLinodeData, linode.ShortName, linodeFile, linode.FullName, linode.HostType, linode.SourceURL},
	{fetchM247, syncM247Data, m247.ShortName, m247File, m247.FullName, m247.HostType, m247.SourceURL},
	{fetchOCI, syncOCIData, oci.ShortName, ociFile, oci.FullName, oci.HostType, oci.SourceURL},
	{fetchOVH, syncOVHData, ovh.ShortName, ovhFile, ovh.FullName, ovh.HostType, ovh.SourceURL},
	{fetchScaleway, syncScalewayData, scaleway.ShortName, scalewayFile, scaleway.FullName, scaleway.HostType, scaleway.SourceURL},
	{fetchVultr, syncVultrData, vultr.ShortName, vultrFile, vultr.FullName, vultr.HostType, vultr.SourceURL},
	{fetchZscaler, syncZscalerData, zscaler.ShortName, zscalerFile, zscaler.FullName, zscaler.HostType, zscaler.SourceURL},
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
