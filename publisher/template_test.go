package publisher_test

import (
	"fmt"
	"testing"

	"github.com/jonhadfield/ip-fetcher/publisher"
)

func TestGenerateReadMeContent(t *testing.T) {
	content, err := publisher.GenerateReadMeContent([]string{"aws", "azure"})
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s\n", content)
}
