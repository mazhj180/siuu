package server

import (
	"encoding/json"
	"io"
	"net/http"
)

func readBody(body io.ReadCloser) (map[string]any, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func addProxy(w http.ResponseWriter, r *http.Request) {

}
