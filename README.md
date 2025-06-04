# ip-fetcher

[![GoDoc](https://godoc.org/github.com/jonhadfield/ip-fetcher?status.svg)](https://pkg.go.dev/github.com/jonhadfield/ip-fetcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/ip-fetcher)](https://goreportcard.com/report/github.com/jonhadfield/ip-fetcher)

## about

ip-fetcher is a go library and cli used to retrieve public ip prefixes from popular cloud and hosting providers.
Please raise an issue if you have any issues or suggestions for new providers.

## supported providers

- <a href="https://www.abuseipdb.com/" target="_blank">AbuseIPDB</a>
- <a href="https://aws.amazon.com/" target="_blank">AWS</a> (Amazon Web Services)
- <a href="https://www.bing.com/webmasters/help/which-crawlers-does-bing-use-8c184ec0" target="_blank">Bingbot</a>
- <a href="https://www.cloudflare.com/" target="_blank">Cloudflare</a>
- <a href="https://www.digitalocean.com/" target="_blank">DigitalOcean</a>
- <a href="https://www.fastly.com/" target="_blank">Fastly</a>
- <a href="https://cloud.google.com/" target="_blank">GCP</a> (Google Cloud Platform)
- <a href="https://www.google.com/" target="_blank">Google</a>
- <a href="https://developers.google.com/search/docs/crawling-indexing/googlebot" target="_blank">Googlebot</a>
- <a href="https://www.maxmind.com" target="_blank">MaxMind GeoIP</a>
- <a href="https://azure.microsoft.com" target="_blank">Microsoft Azure</a>
- <a href="https://www.akamai.com" target="_blank">Akamai</a>
- <a href="https://www.hetzner.com" target="_blank">Hetzner</a>
- <a href="https://www.github.com" target="_blank">GitHub</a>
- <a href="https://www.linode.com" target="_blank">Linode</a>
- <a href="https://www.oracle.com/cloud/" target="_blank">Oracle Cloud Infrastructure</a>
- <a href="https://support.apple.com/en-us/HT212614" target="_blank">iCloud Private Relay</a>
- <a href="https://www.ovhcloud.com" target="_blank">OVHcloud</a>
- <a href="https://www.zscaler.com" target="_blank">Zscaler</a>
- <a href="https://support.virustotal.com" target="_blank">VirusTotal</a>

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
for example:
- output aws prefixes to the console: `ip-fetcher aws --stdout`
- save gcp prefixes to a file: `ip-fetcher gcp --file prefixes.json`

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
