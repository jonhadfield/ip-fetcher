package publisher_test

import (
	"strings"
	"testing"

	"github.com/jonhadfield/ip-fetcher/publisher"
)

func TestGenerateReadMeContent(t *testing.T) {
	_, err := publisher.GenerateReadMeContent([]string{"aws", "azure"})
	if err != nil {
		t.Error(err)
	}
}

// a provider missing from the registry silently produces no row, so assert that
// each registered short name renders one.
func TestGenerateReadMeContentIncludesRegisteredProviders(t *testing.T) {
	cases := []struct {
		shortName string
		wantFile  string
		wantName  string
	}{
		{"akamai", "akamai.txt", "Akamai"},
		{"ahrefs", "ahrefs.json", "AhrefsBot"},
		{"github", "github.txt", "GitHub"},
		{"icloudpr", "icloudpr.csv", "iCloud Private Relay"},
		{"openai", "openai.json", "OpenAI Bots"},
		{"blocklistde", "blocklistde.txt", "Blocklist.de"},
		{"cinsscore", "cinsscore.txt", "CINS Army List"},
		{"cymru", "cymru.json", "Team Cymru Bogons"},
		{"betterstack", "betterstack.txt", "Better Stack"},
		{"checkly", "checkly.json", "Checkly"},
		{"gcore", "gcore.json", "Gcore CDN"},
		{"newrelic", "newrelic.json", "New Relic Synthetics"},
		{"pingdom", "pingdom.json", "Pingdom"},
		{"statuscake", "statuscake.json", "StatusCake"},
		{"zoom", "zoom.txt", "Zoom"},
		{"greensnow", "greensnow.txt", "GreenSnow"},
		{"dshield", "dshield.txt", "DShield Recommended Block List"},
		{"emergingthreats", "emergingthreats.txt", "Emerging Threats Compromised IPs"},
		{"anthropic", "anthropic.json", "Anthropic Crawler Bots"},
		{"applebot", "applebot.json", "Applebot"},
		{"duckduckbot", "duckduckbot.json", "DuckDuckBot"},
		{"perplexitybot", "perplexitybot.json", "PerplexityBot"},
		{"spamhaus", "spamhaus.json", "Spamhaus DROP"},
		{"uptimerobot", "uptimerobot.txt", "UptimeRobot"},
		{"uptrends", "uptrends.json", "Uptrends"},
		{"site24x7", "site24x7.json", "Site24x7"},
		{"grafana", "grafana.json", "Grafana Synthetic Monitoring"},
		{"sentry", "sentry.txt", "Sentry Uptime"},
		{"updown", "updown.json", "updown.io"},
	}

	for _, tc := range cases {
		t.Run(tc.shortName, func(t *testing.T) {
			content, err := publisher.GenerateReadMeContent([]string{tc.shortName})
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(content, tc.wantFile) {
				t.Errorf("generated readme missing file %q for provider %q", tc.wantFile, tc.shortName)
			}

			if !strings.Contains(content, tc.wantName) {
				t.Errorf("generated readme missing name %q for provider %q", tc.wantName, tc.shortName)
			}
		})
	}
}

// an unknown short name must not render a row.
func TestGenerateReadMeContentIgnoresUnknownProvider(t *testing.T) {
	content, err := publisher.GenerateReadMeContent([]string{"nosuchprovider"})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(content, "nosuchprovider") {
		t.Error("generated readme rendered a row for an unregistered provider")
	}
}
