# ip-fetcher

[![GoDoc](https://godoc.org/github.com/jonhadfield/ip-fetcher?status.svg)](https://pkg.go.dev/github.com/jonhadfield/ip-fetcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/ip-fetcher)](https://goreportcard.com/report/github.com/jonhadfield/ip-fetcher)

## about

ip-fetcher is a go library and cli used to retrieve public ip prefixes from popular cloud and hosting providers.
Please raise an issue if you have any issues or suggestions for new providers.

## supported providers

- <a href="https://www.abuseipdb.com/" target="_blank">AbuseIPDB</a>
- <a href="https://api.ahrefs.com/v3/public/crawler-ip-ranges" target="_blank">AhrefsBot</a>
- <a href="https://aws.amazon.com/" target="_blank">AWS</a> (Amazon Web Services)
- <a href="https://www.alibabacloud.com" target="_blank">Alibaba</a>
- <a href="https://claude.com/crawling/bots.json" target="_blank">Anthropic Crawler Bots</a> (ClaudeBot, Claude-User and Claude-SearchBot)
- <a href="https://support.apple.com/en-us/119829" target="_blank">Applebot</a>
- <a href="https://ip-ranges.atlassian.com/" target="_blank">Atlassian</a>
- <a href="https://www.bing.com/webmasters/help/which-crawlers-does-bing-use-8c184ec0" target="_blank">Bingbot</a>
- <a href="https://bunny.net/" target="_blank">Bunny.net</a>
- <a href="https://www.cdn77.com/" target="_blank">CDN77</a>
- <a href="https://www.cloudflare.com/" target="_blank">Cloudflare</a>
- <a href="https://contabo.com/" target="_blank">Contabo</a>
- <a href="https://docs.datadoghq.com/api/latest/ip-ranges/" target="_blank">Datadog</a>
- <a href="https://www.digitalocean.com/" target="_blank">DigitalOcean</a>
- <a href="https://duckduckgo.com/duckduckbot" target="_blank">DuckDuckBot</a>
- <a href="https://www.fastly.com/" target="_blank">Fastly</a>
- <a href="https://fly.io/" target="_blank">Fly.io</a>
- <a href="https://cloud.google.com/" target="_blank">GCP</a> (Google Cloud Platform)
- <a href="https://www.google.com/" target="_blank">Google</a>
- <a href="https://developers.google.com/search/docs/crawling-indexing/googlebot" target="_blank">Googlebot</a>
- <a href="https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot" target="_blank">Google Special Crawlers</a>
- <a href="https://developers.google.com/search/docs/crawling-indexing/verifying-googlebot" target="_blank">Google User-Triggered Fetchers</a>
- <a href="https://www.maxmind.com" target="_blank">MaxMind GeoIP</a>
- <a href="https://azure.microsoft.com" target="_blank">Microsoft Azure</a>
- <a href="https://www.akamai.com" target="_blank">Akamai</a>
- <a href="https://www.hetzner.com" target="_blank">Hetzner</a>
- <a href="https://www.ibm.com/cloud" target="_blank">IBM Cloud</a>
- <a href="https://www.imperva.com" target="_blank">Imperva</a>
- <a href="https://www.github.com" target="_blank">GitHub</a>
- <a href="https://www.leaseweb.com/" target="_blank">Leaseweb</a>
- <a href="https://www.linode.com" target="_blank">Linode</a>
- <a href="https://www.m247.com/" target="_blank">M247</a>
- <a href="https://www.scaleway.com/" target="_blank">Scaleway</a>
- <a href="https://www.spamhaus.org/blocklists/do-not-route-or-peer/" target="_blank">Spamhaus DROP</a>
- <a href="https://docs.stripe.com/ips" target="_blank">Stripe</a>
- <a href="https://www.tencentcloud.com/" target="_blank">Tencent Cloud</a>
- <a href="https://uptimerobot.com/help/locations/" target="_blank">UptimeRobot</a>
- <a href="https://platform.openai.com/docs/bots" target="_blank">OpenAI Bots</a> (GPTBot, OAI-SearchBot and ChatGPT-User)
- <a href="https://www.oracle.com/cloud/" target="_blank">Oracle Cloud Infrastructure</a>
- <a href="https://support.apple.com/en-us/HT212614" target="_blank">iCloud Private Relay</a>
- <a href="https://www.ovhcloud.com" target="_blank">OVHcloud</a>
- <a href="https://www.perplexity.com/perplexitybot.json" target="_blank">PerplexityBot</a>
- <a href="https://render.com/" target="_blank">Render</a>
- <a href="https://www.vultr.com" target="_blank">Vultr</a>
- <a href="https://www.zscaler.com" target="_blank">Zscaler</a>
- <a href="https://github.com/jonhadfield/ip-fetcher" target="_blank">Custom URL</a>

## CLI

### install

Download the latest release [here](https://github.com/jonhadfield/ip-fetcher/releases) and then install:

```bash
install <ip-fetcher binary> /usr/local/bin/ip-fetcher
```
_use: `sudo install` if on linux_

### run

```
ip-fetcher <provider> <options>
```

Use `ip-fetcher --help` to list the providers, and `ip-fetcher <provider> --help`
for a provider's options.

#### common options

| option | description |
| ------ | ----------- |
| `--stdout`, `-s` | write the prefixes to stdout |
| `--Path`, `-p` | write the prefixes to this path; if it is an existing directory then a provider specific default filename is used |
| `--format`, `-f` | output format, on providers that support it: `json` (default), `yaml`, `lines`, `csv` |
| `--lines` | shorthand for `--format lines` (newline separated prefixes) |

At least one of `--stdout` and `--Path` must be given.

#### examples

```bash
# output aws prefixes to the console
ip-fetcher aws --stdout

# save gcp prefixes to a file
ip-fetcher gcp --Path prefixes.json

# write cloudflare's ipv4 prefixes to a directory, using the default filename
ip-fetcher cloudflare -4 --Path /tmp

# newline separated prefixes
ip-fetcher fastly --format lines --stdout

# read prefixes from one or more arbitrary URLs
ip-fetcher url --stdout https://example.com/ips.txt

# providers requiring credentials
ip-fetcher abuseipdb --key <api key> --confidence 90 --stdout
ip-fetcher geoip --key <license key> --Path /tmp --format mmdb
```

#### publish

`ip-fetcher publish` fetches a subset of the providers above (those wired into
`publisher/`) and commits the results, plus a generated README, to a git
repository. It is configured with:

| environment variable | description |
| -------------------- | ----------- |
| `GITHUB_PUBLISH_URL` | the repository to publish to |
| `GITHUB_TOKEN` | a token with write access to that repository |

#### logging

Set `PREFIX_FETCHER_LOG` to change log verbosity, e.g.
`PREFIX_FETCHER_LOG=debug ip-fetcher aws --stdout`.

## API

The following example uses the GCP (Google Cloud Platform) provider.

### installation
```
go get github.com/jonhadfield/ip-fetcher/providers/gcp
```
### basic usage
```
package main

import (
    "fmt"
    "github.com/jonhadfield/ip-fetcher/providers/gcp"
)

func main() {
    g := gcp.New()         // initialise client
    doc, err := g.Fetch()  // fetch prefixes document
    if err != nil {
        panic(err)
    }

    for _, p := range doc.IPv6Prefixes {
        fmt.Printf("%s %s %s\n", p.IPv6Prefix.String(), p.Service, p.Scope)
    }
}
```
