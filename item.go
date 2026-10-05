package arcgis

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ArcGISOnline is the portal that hosts public ArcGIS Online items.
const ArcGISOnline = "https://www.arcgis.com"

var itemIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// ItemURL resolves a portal item ID to the service URL it currently points at.
// Hosted datasets are often republished under new service names while keeping
// their item ID, so the ID is the stable handle.
func ItemURL(ctx context.Context, portalURL, itemID string, opts ...ClientOption) (string, error) {
	if !itemIDPattern.MatchString(itemID) {
		return "", fmt.Errorf("arcgis item %q: not a 32-character hex item ID", itemID)
	}
	c := NewClient(strings.TrimRight(portalURL, "/"), opts...)
	endpoint := c.baseURL + "/sharing/rest/content/items/" + itemID
	var item struct {
		URL string `json:"url"`
	}
	if err := c.get(ctx, endpoint, url.Values{"f": {string(FormatJSON)}}, &item); err != nil {
		return "", fmt.Errorf("arcgis item %s: %w", itemID, err)
	}
	if item.URL == "" {
		return "", fmt.Errorf("arcgis item %s: %w", itemID, errNoItemURL)
	}
	return item.URL, nil
}

var errNoItemURL = errors.New("item has no service URL")
