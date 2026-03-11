package crmapi

import "context"

func (c *Client) ProductsActive(ctx context.Context) (map[string]CategoryBucket, error) {
	var raw map[string]struct {
		Title    string `json:"title"`
		Products map[string]struct {
			Title      string `json:"title"`
			PriceMinor int64  `json:"price_minor"`
			PriceUSD   *int64 `json:"price_usd"`
		} `json:"products"`
	}

	if err := c.get(ctx, "/api/products/active", nil, true, &raw); err != nil {
		return nil, err
	}

	out := make(map[string]CategoryBucket, len(raw))
	for catKey, bucket := range raw {
		products := make(map[string]ProductEntry, len(bucket.Products))
		for pid, p := range bucket.Products {
			products[pid] = ProductEntry{
				Title:      p.Title,
				PriceMinor: p.PriceMinor,
				PriceUSD:   p.PriceUSD,
			}
		}

		out[catKey] = CategoryBucket{
			Title:    bucket.Title,
			Products: products,
		}
	}

	return out, nil
}
