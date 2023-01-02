package fetchers

import "net/http"

type WebFileFetcher interface {
	FetchData() ([]byte, http.Header, int, error)
}
