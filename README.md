# prefix-fetcher 
[![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/prefix-fetcher)](https://goreportcard.com/report/github.com/jonhadfield/prefix-fetcher)

- [about](#about)
- [currently supported providers](#currently-supported-providers)
- [cli](#cli)

## about

prefix-fetcher is a go library and cli used to download and process public IPs from popular cloud and hosting providers.  
Please raise an issue if you have any issues or suggestions for new providers.  
PRs welcomed.

## currently supported providers

- [AbuseIPDB](https://www.abuseipdb.com/)
- [AWS](https://aws.amazon.com/) (Amazon Web Services)
- [Azure](https://azure.microsoft.com)
- [DigitalOcean](https://www.digitalocean.com/)
- [GCP](https://cloud.google.com/) (Google Cloud Platform)
- [MaxMind GeoIP](https://www.maxmind.com)

## cli

Download the latest release [here](https://github.com/jonhadfield/prefix-fetcher/releases) and then install:

```bash
install <prefix-fetcher binary> /usr/local/bin/prefix-fetcher
```

### run

```bash
prefix-fetcher <options> <provider>
```
