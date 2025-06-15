package publisher_test

import (
	"testing"

	"github.com/jonhadfield/ip-fetcher/publisher"
)

func TestGenerateReadMeContent(t *testing.T) {
	_, err := publisher.GenerateReadMeContent([]string{"aws", "azure"})
	if err != nil {
		t.Error(err)
	}
}
