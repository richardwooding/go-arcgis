package arcgis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is the main entry point for interacting with an ArcGIS Feature Service.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string // optional: ArcGIS token for authenticated services
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

// WithToken sets an ArcGIS authentication token.
func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

// WithTimeout sets a request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient = &http.Client{Timeout: d}
	}
}

// NewClient creates a new ArcGIS Feature Service client.
// baseURL should be the root of the FeatureServer, e.g.:
//   https://citymaps.capetown.gov.za/agsext/rest/services/Theme_Based/Open_Data_Service/FeatureServer
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Layer returns a LayerClient scoped to a specific layer ID.
func (c *Client) Layer(id int) *LayerClient {
	return &LayerClient{client: c, layerID: id}
}

// ServiceInfo fetches metadata about the feature service.
func (c *Client) ServiceInfo(ctx context.Context) (*ServiceInfo, error) {
	endpoint := fmt.Sprintf("%s?f=json", c.baseURL)
	var info ServiceInfo
	if err := c.get(ctx, endpoint, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Query executes a single-page query and returns the raw FeatureSet.
func (c *Client) Query(ctx context.Context, p QueryParams) (*FeatureSet, error) {
	p.defaults()
	endpoint := fmt.Sprintf("%s/%d/query?%s", c.baseURL, p.LayerID, p.toQueryString())
	if c.token != "" {
		endpoint += "&token=" + url.QueryEscape(c.token)
	}
	var fs FeatureSet
	if err := c.get(ctx, endpoint, &fs); err != nil {
		return nil, fmt.Errorf("arcgis query layer %d: %w", p.LayerID, err)
	}
	return &fs, nil
}

// QueryAll paginates through all results and returns every Feature.
func (c *Client) QueryAll(ctx context.Context, p QueryParams) ([]Feature, error) {
	p.defaults()
	var all []Feature
	for {
		fs, err := c.Query(ctx, p)
		if err != nil {
			return nil, err
		}
		all = append(all, fs.Features...)
		if !fs.ExceededTransferLimit || len(fs.Features) == 0 {
			break
		}
		p.ResultOffset += len(fs.Features)
	}
	return all, nil
}

// QueryCount returns the count of features matching the query.
func (c *Client) QueryCount(ctx context.Context, p QueryParams) (int, error) {
	p.defaults()
	p.ReturnCountOnly = true
	endpoint := fmt.Sprintf("%s/%d/query?%s", c.baseURL, p.LayerID, p.toQueryString())
	var result struct {
		Count int `json:"count"`
	}
	if err := c.get(ctx, endpoint, &result); err != nil {
		return 0, err
	}
	return result.Count, nil
}

// QueryIDs returns all object IDs matching the query.
func (c *Client) QueryIDs(ctx context.Context, p QueryParams) ([]int64, error) {
	p.defaults()
	p.ReturnIdsOnly = true
	endpoint := fmt.Sprintf("%s/%d/query?%s", c.baseURL, p.LayerID, p.toQueryString())
	var result struct {
		ObjectIDs []int64 `json:"objectIds"`
	}
	if err := c.get(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return result.ObjectIDs, nil
}

// LayerInfo fetches metadata for a specific layer.
func (c *Client) LayerInfo(ctx context.Context, layerID int) (*LayerInfo, error) {
	endpoint := fmt.Sprintf("%s/%d?f=json", c.baseURL, layerID)
	var info LayerInfo
	if err := c.get(ctx, endpoint, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// --- internal ---

func (c *Client) get(ctx context.Context, endpoint string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- LayerClient ---

// LayerClient is scoped to a single layer and provides the fluent query entry point.
type LayerClient struct {
	client  *Client
	layerID int
}

// Query returns a QueryBuilder pre-scoped to this layer.
func (l *LayerClient) Query() *QueryBuilder {
	return &QueryBuilder{
		client: l.client,
		params: QueryParams{LayerID: l.layerID},
	}
}

// Info fetches metadata for this layer.
func (l *LayerClient) Info(ctx context.Context) (*LayerInfo, error) {
	return l.client.LayerInfo(ctx, l.layerID)
}
