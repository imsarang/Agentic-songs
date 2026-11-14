package discover

import (
	"context"
	"encoding/json"
	"net/http"
)

func Probe(ctx context.Context, host string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://"+host+"/products.json?limit=1", nil,
	)

	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "shanks-ecommerce")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false
	}

	defer res.Body.Close()

	var body struct {
		Products []json.RawMessage `json:"products"`
	}

	return res.StatusCode == 200 && json.NewDecoder(res.Body).Decode(&body) == nil
}
