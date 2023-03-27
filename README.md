# prefix-fetcher 

[![GoDoc](https://godoc.org/github.com/jonhadfield/prefix-fetcher?status.svg)](https://pkg.go.dev/github.com/jonhadfield/prefix-fetcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/prefix-fetcher)](https://goreportcard.com/report/github.com/jonhadfield/prefix-fetcher)

- [about](#about)
- [supported providers](#supported-providers)
- [cli](#cli)
  - [install](#install)	
  - [run](#run)
- [api](#api)
  - [installation](#installation)	
  - [basic usage](#basic-usage)
    - [initialise client](#initialise-client) 
    - [fetch](#fetch)  
    - [output](#output)  
 
## about

prefix-fetcher is a go library and cli used to retrieve public ip prefixes from popular cloud and hosting providers.  
Please raise an issue if you have any issues or suggestions for new providers.  

## supported providers

- <a href="https://www.abuseipdb.com/" target="_blank">AbuseIPDB</a>
- <a href="https://aws.amazon.com/" target="_blank">AWS</a> (Amazon Web Services)
- <a href="https://azure.microsoft.com" target="_blank">Azure</a>
- <a href="https://www.digitalocean.com/" target="_blank">DigitalOcean</a>
- <a href="https://cloud.google.com/" target="_blank">GCP</a> (Google Cloud Platform)
- <a href="https://www.maxmind.com" target="_blank">MaxMind GeoIP</a>

## CLI

### install

Download the latest release [here](https://github.com/jonhadfield/prefix-fetcher/releases) and then install:

```bash
install <prefix-fetcher binary> /usr/local/bin/prefix-fetcher
```
_use: `sudo install` if on linux_

### run

```
prefix-fetcher <provider> <options>
```  
for example:  
- output aws prefixes to the console: `prefix-fetcher aws --stdout`  
- save gcp prefixes to a file: `prefix-fetcher gcp --file prefixes.json` 

## API

The following example uses the GCP (Google Cloud Platform) provider. 

### installation
```
go get github.com/jonhadfield/prefix-fetcher/gcp
```
### basic usage
```
package main

import (
	"fmt"
	"github.com/jonhadfield/prefix-fetcher/gcp"
)

func main() {
	g := gcp.New() // initialise client
	doc, err := g.Fetch() // fetch prefixes document
	if err != nil {
		panic(err)
	}

	for _, p := range doc.IPv6Prefixes {
		fmt.Printf("%s %s %s\n", p.IPv6Prefix.String(), p.Service, p.Scope)
	}
}
```
