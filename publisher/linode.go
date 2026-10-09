package publisher

import (
	"encoding/json"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/linode"
)

const linodeFile = "linode.json"

func fetchLinode() ([]byte, error) {
	a := linode.New()

	data, err := a.Fetch()
	if err != nil {
		return nil, err
	}

	records := make([]map[string]any, 0, len(data.Records))
	for _, record := range data.Records {
		records = append(records, map[string]any{
			"prefix":     record.Prefix.String(),
			"alpha2code": record.Alpha2Code,
			"region":     record.Region,
			"city":       record.City,
			"postalCode": record.PostalCode,
		})
	}

	intermediate := map[string]any{
		"lastModified": data.LastModified,
		"etag":         data.ETag,
		"records":      records,
	}

	return json.MarshalIndent(intermediate, "", "  ")
}

func syncLinodeData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(linodeFile, data, wt, fs)
}
