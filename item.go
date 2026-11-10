package arcgis

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ArcGISOnline is the portal that hosts public ArcGIS Online items.
const ArcGISOnline = "https://www.arcgis.com"

var itemIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// Item is the portal metadata of a hosted item. Hosted datasets are often
// republished under new service names while keeping their item ID, so the
// item's title and snippet describe the data better than the service name
// embedded in URL does, and Modified says when the publisher last touched it.
type Item struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Snippet     string   `json:"snippet"`
	Description string   `json:"description"` // HTML, as published
	Type        string   `json:"type"`        // e.g. "Feature Service"
	URL         string   `json:"url"`         // the service URL; empty for items without one
	Tags        []string `json:"tags"`
	LicenseInfo string   `json:"licenseInfo"` // HTML, as published
	Modified    int64    `json:"modified"`    // epoch milliseconds
	Created     int64    `json:"created"`     // epoch milliseconds
}

// ModifiedTime returns Modified as a time, or the zero time when unset.
func (i *Item) ModifiedTime() time.Time { return epochMillis(i.Modified) }

// CreatedTime returns Created as a time, or the zero time when unset.
func (i *Item) CreatedTime() time.Time { return epochMillis(i.Created) }

func epochMillis(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}

// ItemInfo fetches a portal item's metadata: title, snippet, description,
// tags, service URL and timestamps. portalURL is the portal root, for example
// [ArcGISOnline]; itemID is the 32-character hex item ID.
func ItemInfo(ctx context.Context, portalURL, itemID string, opts ...ClientOption) (*Item, error) {
	if !itemIDPattern.MatchString(itemID) {
		return nil, fmt.Errorf("arcgis item %q: not a 32-character hex item ID", itemID)
	}
	c := NewClient(strings.TrimRight(portalURL, "/"), opts...)
	endpoint := c.baseURL + "/sharing/rest/content/items/" + itemID
	var item Item
	if err := c.get(ctx, endpoint, url.Values{"f": {string(FormatJSON)}}, &item); err != nil {
		return nil, fmt.Errorf("arcgis item %s: %w", itemID, err)
	}
	if item.ID == "" {
		item.ID = itemID
	}
	return &item, nil
}

// ItemURL resolves a portal item ID to the service URL it currently points at.
// Hosted datasets are often republished under new service names while keeping
// their item ID, so the ID is the stable handle. Use [ItemInfo] when the title
// or other metadata is wanted as well.
func ItemURL(ctx context.Context, portalURL, itemID string, opts ...ClientOption) (string, error) {
	item, err := ItemInfo(ctx, portalURL, itemID, opts...)
	if err != nil {
		return "", err
	}
	if item.URL == "" {
		return "", fmt.Errorf("arcgis item %s: %w", itemID, errNoItemURL)
	}
	return item.URL, nil
}

var errNoItemURL = errors.New("item has no service URL")
