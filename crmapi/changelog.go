package crmapi

import (
	"context"
	"strings"
)

// ServersChangelog returns the changelog entries a client on `currentVersion`
// has NOT received yet, plus the latest known version.
//
// There is no stored "latest version" anywhere in the system — a worker's
// version is probed live over HTTP. The CRM changelog table is the source of
// truth instead: its highest version label IS the latest release, so Latest and
// Versions can never drift apart.
//
// Pass an empty currentVersion when the worker is down and its version could
// not be read: the result then carries Known=false, and an empty Versions means
// "cannot tell", NOT "up to date" — render it accordingly.
//
// Version comparison happens CRM-side and is numeric per segment, so 1.8.10
// correctly sorts above 1.8.9 (a plain string compare would invert them).
func (c *Client) ServersChangelog(ctx context.Context, currentVersion string) (*ChangelogResult, error) {
	query := map[string]string{}
	if v := strings.TrimSpace(currentVersion); v != "" {
		query["version"] = v
	}

	var raw struct {
		Latest   *string `json:"latest"`
		Known    bool    `json:"known"`
		UpToDate bool    `json:"up_to_date"`
		Versions []struct {
			Version    string  `json:"version"`
			ReleasedAt *string `json:"released_at"`
			Items      []struct {
				Text string `json:"text"`
				Kind string `json:"kind"`
			} `json:"items"`
		} `json:"versions"`
	}

	if err := c.get(ctx, "/api/servers/changelog", query, true, &raw); err != nil {
		return nil, err
	}

	versions := make([]ChangelogVersion, 0, len(raw.Versions))
	for _, v := range raw.Versions {
		items := make([]ChangelogItem, 0, len(v.Items))
		for _, it := range v.Items {
			items = append(items, ChangelogItem{Text: it.Text, Kind: it.Kind})
		}
		versions = append(versions, ChangelogVersion{
			Version:    v.Version,
			ReleasedAt: v.ReleasedAt,
			Items:      items,
		})
	}

	return &ChangelogResult{
		Latest:   raw.Latest,
		Known:    raw.Known,
		UpToDate: raw.UpToDate,
		Versions: versions,
	}, nil
}
