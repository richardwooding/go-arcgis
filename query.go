package arcgis

import (
	"context"
	"fmt"
	"strings"
)

// OutputFormat controls the response format from ArcGIS.
type OutputFormat string

const (
	FormatGeoJSON OutputFormat = "geojson"
	FormatJSON    OutputFormat = "json"
	FormatPBF     OutputFormat = "pbf"
)

// QueryParams defines all parameters for an ArcGIS feature query.
// It can be used directly (struct-style) or built via QueryBuilder (fluent-style).
type QueryParams struct {
	LayerID          int
	Where            string
	Fields           []string
	Envelope         *Envelope
	Geometry         *Point
	GeometryType     GeometryType
	SpatialRel       SpatialRel
	OrderByFields    []string
	GroupByFields    []string
	ResultOffset     int
	PageSize         int
	ReturnGeometry   *bool // nil = server default (true)
	ReturnIdsOnly    bool
	ReturnCountOnly  bool
	Format           OutputFormat
}

// defaults applies sensible defaults to unset fields.
func (p *QueryParams) defaults() {
	if p.Where == "" {
		p.Where = "1=1"
	}
	if p.Format == "" {
		p.Format = FormatGeoJSON
	}
	if p.GeometryType == "" && p.Envelope != nil {
		p.GeometryType = GeometryTypeEnvelope
	}
	if p.SpatialRel == "" && (p.Envelope != nil || p.Geometry != nil) {
		p.SpatialRel = SpatialRelIntersects
	}
	if p.PageSize == 0 {
		p.PageSize = 1000
	}
}

// toQueryString converts QueryParams to ArcGIS REST query parameters.
func (p QueryParams) toQueryString() string {
	params := []string{
		fmt.Sprintf("f=%s", p.Format),
		fmt.Sprintf("where=%s", p.Where),
		fmt.Sprintf("resultOffset=%d", p.ResultOffset),
		fmt.Sprintf("resultRecordCount=%d", p.PageSize),
	}

	if len(p.Fields) > 0 {
		params = append(params, fmt.Sprintf("outFields=%s", strings.Join(p.Fields, ",")))
	} else {
		params = append(params, "outFields=*")
	}

	if p.Envelope != nil {
		geom := fmt.Sprintf("%f,%f,%f,%f", p.Envelope.MinX, p.Envelope.MinY, p.Envelope.MaxX, p.Envelope.MaxY)
		params = append(params,
			fmt.Sprintf("geometry=%s", geom),
			fmt.Sprintf("geometryType=%s", p.GeometryType),
			fmt.Sprintf("spatialRel=%s", p.SpatialRel),
		)
	}

	if len(p.OrderByFields) > 0 {
		params = append(params, fmt.Sprintf("orderByFields=%s", strings.Join(p.OrderByFields, ",")))
	}

	if p.ReturnIdsOnly {
		params = append(params, "returnIdsOnly=true")
	}
	if p.ReturnCountOnly {
		params = append(params, "returnCountOnly=true")
	}
	if p.ReturnGeometry != nil && !*p.ReturnGeometry {
		params = append(params, "returnGeometry=false")
	}

	return strings.Join(params, "&")
}

// --- Fluent QueryBuilder ---

// QueryBuilder builds a QueryParams using a fluent chainable API.
type QueryBuilder struct {
	client *Client
	params QueryParams
}

// Where sets the SQL WHERE clause.
func (q *QueryBuilder) Where(clause string) *QueryBuilder {
	q.params.Where = clause
	return q
}

// Fields sets the output fields to return.
func (q *QueryBuilder) Fields(fields ...string) *QueryBuilder {
	q.params.Fields = fields
	return q
}

// WithinEnvelope sets a bounding box spatial filter.
func (q *QueryBuilder) WithinEnvelope(minX, minY, maxX, maxY float64) *QueryBuilder {
	q.params.Envelope = &Envelope{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	return q
}

// OrderBy sets the ORDER BY fields.
func (q *QueryBuilder) OrderBy(fields ...string) *QueryBuilder {
	q.params.OrderByFields = fields
	return q
}

// PageSize sets the number of records per page.
func (q *QueryBuilder) PageSize(n int) *QueryBuilder {
	q.params.PageSize = n
	return q
}

// WithoutGeometry omits geometry from the response (faster for attribute-only queries).
func (q *QueryBuilder) WithoutGeometry() *QueryBuilder {
	f := false
	q.params.ReturnGeometry = &f
	return q
}

// Format sets the response format.
func (q *QueryBuilder) Format(f OutputFormat) *QueryBuilder {
	q.params.Format = f
	return q
}

// From merges a pre-built QueryParams into this builder (useful for CCT named queries).
func (q *QueryBuilder) From(base QueryParams) *QueryBuilder {
	// Preserve the layerID set by Layer(), merge everything else
	layerID := q.params.LayerID
	q.params = base
	q.params.LayerID = layerID
	return q
}

// Params returns the underlying QueryParams for inspection or reuse.
func (q *QueryBuilder) Params() QueryParams {
	return q.params
}

// First fetches only the first page of results.
func (q *QueryBuilder) First(ctx context.Context) (*FeatureSet, error) {
	q.params.defaults()
	return q.client.Query(ctx, q.params)
}

// All fetches all results, handling pagination automatically.
func (q *QueryBuilder) All(ctx context.Context) ([]Feature, error) {
	q.params.defaults()
	return q.client.QueryAll(ctx, q.params)
}

// Count returns only the record count matching the query.
func (q *QueryBuilder) Count(ctx context.Context) (int, error) {
	q.params.defaults()
	q.params.ReturnCountOnly = true
	return q.client.QueryCount(ctx, q.params)
}

// IDs returns only the object IDs matching the query.
func (q *QueryBuilder) IDs(ctx context.Context) ([]int64, error) {
	q.params.defaults()
	q.params.ReturnIdsOnly = true
	return q.client.QueryIDs(ctx, q.params)
}
