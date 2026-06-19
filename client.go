// Package arcgis is a small, dependency-free client for querying ArcGIS
// Feature Services over their REST API.
//
// It offers two interchangeable styles: a struct-based API ([QueryParams] passed
// to [Client.Query]) and a fluent builder ([Client.Layer] → [LayerClient.Query]).
// Pagination, counts, and object-ID-only queries are first-class.
//
//	client := arcgis.NewClient(baseURL)
//	features, err := client.Layer(7).Query().
//		Where("STAGE = 4").
//		Fields("NAME", "STAGE").
//		All(ctx)
package arcgis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// maxQueryStringLen is the encoded-parameter length above which requests are
// sent as POST instead of GET. A query carrying a detailed geometry (e.g. a
// ward or suburb polygon with hundreds of vertices) easily exceeds typical
// server URL-length limits (~2048), which manifests as an HTTP 404. ArcGIS REST
// endpoints accept the same parameters by GET or POST, so we fall back to POST.
const maxQueryStringLen = 1800

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

// WithToken sets an ArcGIS authentication token, appended to every request.
func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

// WithTimeout sets a request timeout on the default HTTP client. It has no
// effect when combined with WithHTTPClient.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient = &http.Client{Timeout: d}
	}
}

// NewClient creates a new ArcGIS Feature Service client. baseURL should be the
// root of the FeatureServer, e.g.:
//
//	https://example.gov/arcgis/rest/services/Theme/Service/FeatureServer
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
	var info ServiceInfo
	if err := c.get(ctx, c.baseURL, url.Values{"f": {"json"}}, &info); err != nil {
		return nil, fmt.Errorf("arcgis service info: %w", err)
	}
	return &info, nil
}

// LayerInfo fetches metadata for a specific layer.
func (c *Client) LayerInfo(ctx context.Context, layerID int) (*LayerInfo, error) {
	endpoint := fmt.Sprintf("%s/%d", c.baseURL, layerID)
	var info LayerInfo
	if err := c.get(ctx, endpoint, url.Values{"f": {"json"}}, &info); err != nil {
		return nil, fmt.Errorf("arcgis layer %d info: %w", layerID, err)
	}
	return &info, nil
}

// Query executes a single-page query and returns the raw FeatureSet.
func (c *Client) Query(ctx context.Context, p QueryParams) (*FeatureSet, error) {
	p.defaults()
	var fs FeatureSet
	if err := c.get(ctx, c.queryEndpoint(p), p.values(), &fs); err != nil {
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
	p.Format = FormatJSON // count responses are Esri JSON only
	var result struct {
		Count int `json:"count"`
	}
	if err := c.get(ctx, c.queryEndpoint(p), p.values(), &result); err != nil {
		return 0, fmt.Errorf("arcgis count layer %d: %w", p.LayerID, err)
	}
	return result.Count, nil
}

// QueryIDs returns all object IDs matching the query.
func (c *Client) QueryIDs(ctx context.Context, p QueryParams) ([]int64, error) {
	p.defaults()
	p.ReturnIDsOnly = true
	p.Format = FormatJSON // ID responses are Esri JSON only
	var result struct {
		ObjectIDs []int64 `json:"objectIds"`
	}
	if err := c.get(ctx, c.queryEndpoint(p), p.values(), &result); err != nil {
		return nil, fmt.Errorf("arcgis ids layer %d: %w", p.LayerID, err)
	}
	return result.ObjectIDs, nil
}

// --- internal ---

func (c *Client) queryEndpoint(p QueryParams) string {
	return fmt.Sprintf("%s/%d/query", c.baseURL, p.LayerID)
}

func (c *Client) get(ctx context.Context, endpoint string, params url.Values, out any) error {
	if c.token != "" {
		params.Set("token", c.token)
	}
	req, err := c.newRequest(ctx, endpoint, params)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(body, 4096))
	}

	// ArcGIS often reports failures with HTTP 200 and an error envelope.
	var env errorEnvelope
	if json.Unmarshal(body, &env) == nil && env.Error != nil {
		return env.Error
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// newRequest builds a GET request, falling back to a POST with a form-encoded
// body when the query string is large enough to risk exceeding server
// URL-length limits (e.g. a query carrying a detailed polygon geometry).
func (c *Client) newRequest(ctx context.Context, endpoint string, params url.Values) (*http.Request, error) {
	encoded := params.Encode()
	if len(encoded) <= maxQueryStringLen {
		return http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+encoded, http.NoBody)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// truncate renders up to n bytes of body for inclusion in an error message.
func truncate(body []byte, n int) string {
	if len(body) > n {
		return string(body[:n])
	}
	return string(body)
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
