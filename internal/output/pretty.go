package output

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func PrettyJSON(data []byte) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}

func PrettyPrintJSON(data []byte) error {
	pj, err := PrettyJSON(data)
	if err != nil {
		return fmt.Errorf("error pretty printing JSON: %w", err)
	}

	fmt.Println(pj)

	return nil
}
