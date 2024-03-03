package publisher

import (
	"fmt"
	"testing"
)

func TestGenerateReadMeContent(t *testing.T) {
	content, err := generateReadMeContent([]string{"aws", "azure"})
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s\n", content)
}
