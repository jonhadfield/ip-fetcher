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

- [AbuseIPDB](https://www.abuseipdb.com/)
- [AWS](https://aws.amazon.com/) (Amazon Web Services)
- [Azure](https://azure.microsoft.com)
- [DigitalOcean](https://www.digitalocean.com/)
- [GCP](https://cloud.google.com/) (Google Cloud Platform)
- [MaxMind GeoIP](https://www.maxmind.com)

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
	g := gcp.New()
	doc, err := g.Fetch()
	if err != nil {
		panic(err)
	}

	for _, p := range doc.IPv6Prefixes {
		fmt.Printf("%s %s %s\n", p.IPv6Prefix.String(), p.Service, p.Scope)
	}
}
```
