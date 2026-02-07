package publisher

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"

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
	rgb, err := fs.Open(linodeFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(linodeFile, "up to date", upToDate)
	}

	if err = createFile(fs, linodeFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(linodeFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update linode data")
}
