package arcgis

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// OutputFormat controls the response format from ArcGIS.
type OutputFormat string

const (
	// FormatGeoJSON requests RFC 7946 GeoJSON. Each feature carries its
	// attributes under "properties".
	FormatGeoJSON OutputFormat = "geojson"
	// FormatJSON requests Esri JSON. Each feature carries its attributes
	// under "attributes".
	FormatJSON OutputFormat = "json"
	// FormatPBF requests the Protocol Buffer encoding. This package decodes
	// JSON responses only; use FormatPBF only with a custom decoder.
	FormatPBF OutputFormat = "pbf"
)

// QueryParams defines all parameters for an ArcGIS feature query.
// It can be used directly (struct-style) or built via QueryBuilder (fluent-style).
//
// The zero value is usable: defaults are applied for an unset Where clause
// ("1=1"), Format (GeoJSON), and PageSize (1000).
type QueryParams struct {
	LayerID         int
	Where           string
	Fields          []string
	Envelope        *Envelope
	Geometry        *Point
	GeometryType    GeometryType
	SpatialRel      SpatialRel
	OrderByFields   []string
	GroupByFields   []string
	ResultOffset    int
	PageSize        int
	ReturnGeometry  *bool // nil = server default (true)
	ReturnIDsOnly   bool
	ReturnCountOnly bool
	// ReturnDistinctValues requests only distinct values for the selected
	// Fields. Typically combined with Fields (and often OrderByFields) to
	// enumerate the values present in one or more columns.
	ReturnDistinctValues bool
	Format               OutputFormat
}

// defaults applies sensible defaults to unset fields.
func (p *QueryParams) defaults() {
	if p.Where == "" {
		p.Where = "1=1"
	}
	if p.Format == "" {
		p.Format = FormatGeoJSON
	}
	if p.GeometryType == "" {
		switch {
		case p.Envelope != nil:
			p.GeometryType = GeometryTypeEnvelope
		case p.Geometry != nil:
			p.GeometryType = GeometryTypePoint
		}
	}
	if p.SpatialRel == "" && (p.Envelope != nil || p.Geometry != nil) {
		p.SpatialRel = SpatialRelIntersects
	}
	if p.PageSize == 0 {
		p.PageSize = 1000
	}
}

// values converts QueryParams to escaped ArcGIS REST query parameters.
func (p QueryParams) values() url.Values {
	v := url.Values{}
	v.Set("f", string(p.Format))
	v.Set("where", p.Where)
	v.Set("resultOffset", strconv.Itoa(p.ResultOffset))
	v.Set("resultRecordCount", strconv.Itoa(p.PageSize))

	if len(p.Fields) > 0 {
		v.Set("outFields", strings.Join(p.Fields, ","))
	} else {
		v.Set("outFields", "*")
	}

	switch {
	case p.Envelope != nil:
		v.Set("geometry", p.Envelope.bbox())
		v.Set("geometryType", string(p.GeometryType))
		v.Set("spatialRel", string(p.SpatialRel))
	case p.Geometry != nil:
		v.Set("geometry", p.Geometry.coords())
		v.Set("geometryType", string(p.GeometryType))
		v.Set("spatialRel", string(p.SpatialRel))
	}

	if len(p.OrderByFields) > 0 {
		v.Set("orderByFields", strings.Join(p.OrderByFields, ","))
	}
	if len(p.GroupByFields) > 0 {
		v.Set("groupByFieldsForStatistics", strings.Join(p.GroupByFields, ","))
	}

	if p.ReturnIDsOnly {
		v.Set("returnIdsOnly", "true")
	}
	if p.ReturnCountOnly {
		v.Set("returnCountOnly", "true")
	}
	if p.ReturnDistinctValues {
		v.Set("returnDistinctValues", "true")
	}
	if p.ReturnGeometry != nil && !*p.ReturnGeometry {
		v.Set("returnGeometry", "false")
	}

	return v
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

// WithinEnvelope sets a bounding-box spatial filter.
func (q *QueryBuilder) WithinEnvelope(minX, minY, maxX, maxY float64) *QueryBuilder {
	q.params.Envelope = &Envelope{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	return q
}

// IntersectsPoint sets a point spatial filter using the default
// "intersects" relationship.
func (q *QueryBuilder) IntersectsPoint(x, y float64) *QueryBuilder {
	q.params.Geometry = &Point{X: x, Y: y}
	return q
}

// SpatialRel overrides the spatial relationship applied to the geometry filter.
func (q *QueryBuilder) SpatialRel(rel SpatialRel) *QueryBuilder {
	q.params.SpatialRel = rel
	return q
}

// OrderBy sets the ORDER BY fields.
func (q *QueryBuilder) OrderBy(fields ...string) *QueryBuilder {
	q.params.OrderByFields = fields
	return q
}

// GroupBy sets the GROUP BY fields (used with statistics queries).
func (q *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	q.params.GroupByFields = fields
	return q
}

// DistinctValues requests only distinct values for the selected Fields.
func (q *QueryBuilder) DistinctValues() *QueryBuilder {
	q.params.ReturnDistinctValues = true
	return q
}

// Offset sets the starting record offset for pagination.
func (q *QueryBuilder) Offset(n int) *QueryBuilder {
	q.params.ResultOffset = n
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

// From merges a pre-built QueryParams into this builder (useful for named queries).
// The layer ID set by Layer() is preserved.
func (q *QueryBuilder) From(base QueryParams) *QueryBuilder {
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
	return q.client.Query(ctx, q.params)
}

// All fetches all results, handling pagination automatically.
func (q *QueryBuilder) All(ctx context.Context) ([]Feature, error) {
	return q.client.QueryAll(ctx, q.params)
}

// Count returns only the record count matching the query.
func (q *QueryBuilder) Count(ctx context.Context) (int, error) {
	return q.client.QueryCount(ctx, q.params)
}

// IDs returns only the object IDs matching the query.
func (q *QueryBuilder) IDs(ctx context.Context) ([]int64, error) {
	return q.client.QueryIDs(ctx, q.params)
}
