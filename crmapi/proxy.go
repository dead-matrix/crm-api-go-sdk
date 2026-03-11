package crmapi

import (
	"context"
	"fmt"
)

func (c *Client) ProxyCheck(ctx context.Context, userID int64) (*ProxyCheckResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw struct {
		Checked int64 `json:"checked"`
		Valid   int64 `json:"valid"`
		Invalid int64 `json:"invalid"`
		Results []struct {
			Proxy    string  `json:"proxy"`
			Valid    bool    `json:"valid"`
			RUError  *string `json:"ru_error"`
			Location *string `json:"location"`
		} `json:"results"`
	}

	if err := c.post(ctx, "/api/proxy/check", query, true, nil, &raw); err != nil {
		return nil, err
	}

	results := make([]ProxyCheckItem, 0, len(raw.Results))
	for _, r := range raw.Results {
		results = append(results, ProxyCheckItem{
			Proxy:    r.Proxy,
			Valid:    r.Valid,
			RUError:  r.RUError,
			Location: r.Location,
		})
	}

	return &ProxyCheckResult{
		Checked: raw.Checked,
		Valid:   raw.Valid,
		Invalid: raw.Invalid,
		Results: results,
	}, nil
}

func (c *Client) ProxyList(ctx context.Context, userID int64) ([]ProxyItem, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw []struct {
		Type     *string `json:"type"`
		IP       *string `json:"ip"`
		Port     *int64  `json:"port"`
		Login    *string `json:"login"`
		Password *string `json:"password"`
		Valid    bool    `json:"valid"`
		Location *string `json:"location"`
	}

	if err := c.get(ctx, "/api/proxy/list", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]ProxyItem, 0, len(raw))
	for _, r := range raw {
		items = append(items, ProxyItem{
			Type:     r.Type,
			IP:       r.IP,
			Port:     r.Port,
			Login:    r.Login,
			Password: r.Password,
			Valid:    r.Valid,
			Location: r.Location,
		})
	}

	return items, nil
}
