package publisher

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/jonhadfield/ip-fetcher/providers/m247"
	"github.com/jonhadfield/ip-fetcher/providers/scaleway"
	"github.com/jonhadfield/ip-fetcher/providers/vultr"

	"github.com/jonhadfield/ip-fetcher/providers/akamai"
	"github.com/jonhadfield/ip-fetcher/providers/alibaba"
	"github.com/jonhadfield/ip-fetcher/providers/openai"
	"github.com/jonhadfield/ip-fetcher/providers/ovh"

	"github.com/jonhadfield/ip-fetcher/providers/hetzner"
	"github.com/jonhadfield/ip-fetcher/providers/ibmcloud"
	"github.com/jonhadfield/ip-fetcher/providers/tencent"
	"github.com/jonhadfield/ip-fetcher/providers/zscaler"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ahrefs"
	"github.com/jonhadfield/ip-fetcher/providers/anthropic"
	"github.com/jonhadfield/ip-fetcher/providers/applebot"
	"github.com/jonhadfield/ip-fetcher/providers/atlassian"
	"github.com/jonhadfield/ip-fetcher/providers/aws"
	"github.com/jonhadfield/ip-fetcher/providers/azure"
	"github.com/jonhadfield/ip-fetcher/providers/blocklistde"
	"github.com/jonhadfield/ip-fetcher/providers/bunny"
	"github.com/jonhadfield/ip-fetcher/providers/cdn77"
	"github.com/jonhadfield/ip-fetcher/providers/cinsscore"
	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
	"github.com/jonhadfield/ip-fetcher/providers/contabo"
	"github.com/jonhadfield/ip-fetcher/providers/datadog"
	"github.com/jonhadfield/ip-fetcher/providers/dshield"
	"github.com/jonhadfield/ip-fetcher/providers/duckduckbot"
	"github.com/jonhadfield/ip-fetcher/providers/emergingthreats"
	"github.com/jonhadfield/ip-fetcher/providers/fastly"
	"github.com/jonhadfield/ip-fetcher/providers/flyio"
	"github.com/jonhadfield/ip-fetcher/providers/gcp"
	"github.com/jonhadfield/ip-fetcher/providers/github"
	"github.com/jonhadfield/ip-fetcher/providers/google"
	"github.com/jonhadfield/ip-fetcher/providers/googlebot"
	"github.com/jonhadfield/ip-fetcher/providers/googlesc"
	"github.com/jonhadfield/ip-fetcher/providers/googleutf"
	"github.com/jonhadfield/ip-fetcher/providers/icloudpr"
	"github.com/jonhadfield/ip-fetcher/providers/imperva"
	"github.com/jonhadfield/ip-fetcher/providers/leaseweb"
	"github.com/jonhadfield/ip-fetcher/providers/linode"
	"github.com/jonhadfield/ip-fetcher/providers/oci"
	"github.com/jonhadfield/ip-fetcher/providers/perplexitybot"
	"github.com/jonhadfield/ip-fetcher/providers/render"
	"github.com/jonhadfield/ip-fetcher/providers/spamhaus"
	"github.com/jonhadfield/ip-fetcher/providers/stripe"
	"github.com/jonhadfield/ip-fetcher/providers/uptimerobot"
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
	{fetchAhrefs, syncAhrefsData, ahrefs.ShortName, ahrefsFile, ahrefs.FullName, ahrefs.HostType, ahrefs.SourceURL},
	{fetchAkamai, syncAkamaiData, akamai.ShortName, akamaiFile, akamai.FullName, akamai.HostType, akamai.SourceURL},
	{fetchAlibaba, syncAlibabaData, alibaba.ShortName, alibabaFile, alibaba.FullName, alibaba.HostType, alibaba.SourceURL},
	{fetchAnthropic, syncAnthropicData, anthropic.ShortName, anthropicFile, anthropic.FullName, anthropic.HostType, anthropic.SourceURL},
	{fetchApplebot, syncApplebotData, applebot.ShortName, applebotFile, applebot.FullName, applebot.HostType, applebot.SourceURL},
	{fetchAtlassian, syncAtlassianData, atlassian.ShortName, atlassianFile, atlassian.FullName, atlassian.HostType, atlassian.SourceURL},
	{fetchAWS, syncAWSData, aws.ShortName, awsFile, aws.FullName, aws.HostType, aws.SourceURL},
	{fetchAzure, syncAzureData, azure.ShortName, azureFile, azure.FullName, azure.HostType, azure.InitialURL},
	{fetchBlocklistde, syncBlocklistdeData, blocklistde.ShortName, blocklistdeFile, blocklistde.FullName, blocklistde.HostType, blocklistde.SourceURL},
	{fetchBunny, syncBunnyData, bunny.ShortName, bunnyFile, bunny.FullName, bunny.HostType, bunny.SourceURL},
	{fetchCDN77, syncCDN77Data, cdn77.ShortName, cdn77File, cdn77.FullName, cdn77.HostType, cdn77.SourceURL},
	{fetchCinsscore, syncCinsscoreData, cinsscore.ShortName, cinsscoreFile, cinsscore.FullName, cinsscore.HostType, cinsscore.SourceURL},
	{fetchCloudflare, syncCloudflareData, cloudflare.ShortName, cloudflareFile, cloudflare.FullName, cloudflare.HostType, cloudflare.SourceURL},
	{fetchContabo, syncContaboData, contabo.ShortName, contaboFile, contabo.FullName, contabo.HostType, contabo.SourceURL},
	{fetchDatadog, syncDatadogData, datadog.ShortName, datadogFile, datadog.FullName, datadog.HostType, datadog.SourceURL},
	{fetchDshield, syncDshieldData, dshield.ShortName, dshieldFile, dshield.FullName, dshield.HostType, dshield.SourceURL},
	{fetchDuckduckbot, syncDuckduckbotData, duckduckbot.ShortName, duckduckbotFile, duckduckbot.FullName, duckduckbot.HostType, duckduckbot.SourceURL},
	{fetchEmergingthreats, syncEmergingthreatsData, emergingthreats.ShortName, emergingthreatsFile, emergingthreats.FullName, emergingthreats.HostType, emergingthreats.SourceURL},
	{fetchFastly, syncFastlyData, fastly.ShortName, fastlyFile, fastly.FullName, fastly.HostType, fastly.SourceURL},
	{fetchFlyio, syncFlyioData, flyio.ShortName, flyioFile, flyio.FullName, flyio.HostType, flyio.SourceURL},
	{fetchGCP, syncGCPData, gcp.ShortName, gcpFile, gcp.FullName, gcp.HostType, gcp.SourceURL},
	{fetchGitHub, syncGitHubData, github.ShortName, githubFile, github.FullName, github.HostType, github.SourceURL},
	{fetchGoogle, syncGoogleData, google.ShortName, googleFile, google.FullName, google.HostType, google.SourceURL},
	{fetchGooglebot, syncGooglebotData, googlebot.ShortName, googlebotFile, googlebot.FullName, googlebot.HostType, googlebot.SourceURL},
	{fetchGoogleSC, syncGoogleSCData, googlesc.ShortName, googlescFile, googlesc.FullName, googlesc.HostType, googlesc.SourceURL},
	{fetchGoogleUTF, syncGoogleUTFData, googleutf.ShortName, googleutfFile, googleutf.FullName, googleutf.HostType, googleutf.SourceURL},
	{fetchHetzner, syncHetznerData, hetzner.ShortName, hetznerFile, hetzner.FullName, hetzner.HostType, hetzner.SourceURL},
	{fetchIBMCloud, syncIBMCloudData, ibmcloud.ShortName, ibmcloudFile, ibmcloud.FullName, ibmcloud.HostType, ibmcloud.SourceURL},
	{fetchICloudPR, syncICloudPRData, icloudpr.ShortName, icloudprFile, icloudpr.FullName, icloudpr.HostType, icloudpr.SourceURL},
	{fetchImperva, syncImpervaData, imperva.ShortName, impervaFile, imperva.FullName, imperva.HostType, imperva.SourceURL},
	{fetchLeaseweb, syncLeasewebData, leaseweb.ShortName, leasewebFile, leaseweb.FullName, leaseweb.HostType, leaseweb.SourceURL},
	{fetchLinode, syncLinodeData, linode.ShortName, linodeFile, linode.FullName, linode.HostType, linode.SourceURL},
	{fetchM247, syncM247Data, m247.ShortName, m247File, m247.FullName, m247.HostType, m247.SourceURL},
	{fetchOCI, syncOCIData, oci.ShortName, ociFile, oci.FullName, oci.HostType, oci.SourceURL},
	{fetchOpenAI, syncOpenAIData, openai.ShortName, openaiFile, openai.FullName, openai.HostType, openai.SourceURL},
	{fetchOVH, syncOVHData, ovh.ShortName, ovhFile, ovh.FullName, ovh.HostType, ovh.SourceURL},
	{fetchPerplexitybot, syncPerplexitybotData, perplexitybot.ShortName, perplexitybotFile, perplexitybot.FullName, perplexitybot.HostType, perplexitybot.SourceURL},
	{fetchRender, syncRenderData, render.ShortName, renderFile, render.FullName, render.HostType, render.SourceURL},
	{fetchScaleway, syncScalewayData, scaleway.ShortName, scalewayFile, scaleway.FullName, scaleway.HostType, scaleway.SourceURL},
	{fetchSpamhaus, syncSpamhausData, spamhaus.ShortName, spamhausFile, spamhaus.FullName, spamhaus.HostType, spamhaus.SourceURL},
	{fetchStripe, syncStripeData, stripe.ShortName, stripeFile, stripe.FullName, stripe.HostType, stripe.SourceURL},
	{fetchTencent, syncTencentData, tencent.ShortName, tencentFile, tencent.FullName, tencent.HostType, tencent.SourceURL},
	{fetchUptimerobot, syncUptimerobotData, uptimerobot.ShortName, uptimerobotFile, uptimerobot.FullName, uptimerobot.HostType, uptimerobot.SourceURL},
	{fetchVultr, syncVultrData, vultr.ShortName, vultrFile, vultr.FullName, vultr.HostType, vultr.SourceURL},
	{fetchZscaler, syncZscalerData, zscaler.ShortName, zscalerFile, zscaler.FullName, zscaler.HostType, zscaler.SourceURL},
}

func GenerateReadMeContent(included []string) (string, error) {
	rows := strings.Builder{}

	for _, inc := range included {
		for _, provider := range providers {
			if inc == provider.ShortName {
				fmt.Fprintf(
					&rows,
					"| [%s](%s)  | %s |  %s | [source](%s) |  \r\n",
					provider.File,
					provider.File,
					provider.FullName,
					provider.HostType,
					provider.SourceURL,
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
