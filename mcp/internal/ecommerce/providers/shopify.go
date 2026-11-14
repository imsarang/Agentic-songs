package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"shanks/internal/ecommerce/config"
	"shanks/internal/ecommerce/discover"
)

const (
	httpTimeout    = 20 * time.Second
	shopofyBaseUrl = ""
	searchUrl      = "/en/search/suggest.json"
)

type Product struct {
	Host   string
	Title  string
	Handle string
	URL    string
	Price  string
}

type HttpClient struct {
	http *http.Client
	live []config.Host
}

func NewHTTPClient() *HttpClient {
	return &HttpClient{
		http: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

func (c *HttpClient) RelevantDomains(ctx context.Context, hosts []config.Host) {
	var mu sync.Mutex
	g, ctx := errgroup.WithContext(ctx)

	for _, host := range hosts {
		host := host
		g.Go(func() error {
			pctx, cancel := context.WithTimeout(
				ctx, 3*time.Second,
			)

			defer cancel()

			if !discover.Probe(pctx, host.Name) {
				return nil
			}

			mu.Lock()
			c.live = append(c.live, host)
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()
}

func (c *HttpClient) SearchProduct(ctx context.Context, group, query string) ([]Product, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var mu sync.Mutex

	var found []Product

	g, ctx := errgroup.WithContext(ctx)

	for _, host := range c.live {
		if group != "" && host.Group != group {
			continue
		}

		host := host
		g.Go(func() error {
			rows, err := c.suggest(ctx, host.Name, query)
			if err != nil {
				return nil
			}

			mu.Lock()
			found = append(found, rows...)
			mu.Unlock()
			return nil
		})
	}

	return found, g.Wait()
}

func (c *HttpClient) suggest(ctx context.Context, host, query string) ([]Product, error) {
	u := "https://" + host + "/search/suggest.json?resources[type]=product&resources[limit]=8&q=" + url.QueryEscape(query)

	var page struct {
		Resources struct {
			Results struct {
				Products []struct {
					Title  string `json:"title"`
					Handle string `json:"handle"`
					URL    string `json:"url"`
					Price  string `json:"price"`
				} `json:"products"`
			} `json:"results"`
		} `json:"resources"`
	}

	if err := c.NetworkCall(ctx, http.MethodGet, u, nil, &page); err != nil {
		return nil, err
	}
	var rows []Product
	for _, row := range page.Resources.Results.Products {
		link := row.URL
		if strings.HasPrefix(link, "/") {
			link = "https://" + host + link
		}
		rows = append(rows, Product{
			Host: host, Title: row.Title, Handle: row.Handle, URL: link, Price: row.Price,
		})
	}
	return rows, nil
}

func (c *HttpClient) NetworkCall(ctx context.Context, method, rawURL string, in, out any) error {
	var payload []byte

	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = b
	}

	var last error
	for attempt := 0; attempt < 2; attempt++ {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "shanks-ecommerce")

		res, err := c.http.Do(req)
		if err != nil {
			last = err
			continue
		}

		b, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			last = fmt.Errorf("%s %s: %s", method, rawURL, res.Status)
			continue
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(b, out)
	}
	if last == nil {
		last = errors.New("request failed")
	}
	return last
}
