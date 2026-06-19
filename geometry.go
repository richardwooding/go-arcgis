package arcgis

import (
	"strconv"
	"strings"
)

// Polygon is a polygon spatial filter, expressed as one or more linear rings of
// [x, y] coordinates in the layer's coordinate system (typically WGS84
// longitude/latitude). Per the Esri convention an exterior ring is clockwise
// and holes are counter-clockwise, but ArcGIS query filters tolerate either.
type Polygon struct {
	Rings [][][]float64
}

// esriJSON renders the polygon as the Esri JSON geometry string ArcGIS accepts
// for a polygon filter: {"rings":[[[x,y],...],...]}.
func (p Polygon) esriJSON() string {
	var b strings.Builder
	b.WriteString(`{"rings":[`)
	for i, ring := range p.Rings {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('[')
		for j, pt := range ring {
			if j > 0 {
				b.WriteByte(',')
			}
			x, y := 0.0, 0.0
			if len(pt) > 0 {
				x = pt[0]
			}
			if len(pt) > 1 {
				y = pt[1]
			}
			b.WriteByte('[')
			b.WriteString(formatFloat(x))
			b.WriteByte(',')
			b.WriteString(formatFloat(y))
			b.WriteByte(']')
		}
		b.WriteByte(']')
	}
	b.WriteString(`]}`)
	return b.String()
}

// Envelope is a bounding-box spatial filter, expressed in the layer's
// coordinate system (typically WGS84 longitude/latitude).
type Envelope struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// bbox renders the envelope as the comma-delimited "minx,miny,maxx,maxy"
// string ArcGIS accepts for an envelope geometry.
func (e Envelope) bbox() string {
	return strings.Join([]string{
		formatFloat(e.MinX), formatFloat(e.MinY),
		formatFloat(e.MaxX), formatFloat(e.MaxY),
	}, ",")
}

// Point is a single coordinate.
type Point struct {
	X float64
	Y float64
}

// coords renders the point as the comma-delimited "x,y" string ArcGIS accepts
// for a point geometry.
func (p Point) coords() string {
	return formatFloat(p.X) + "," + formatFloat(p.Y)
}

// formatFloat renders a coordinate without scientific notation or trailing
// zero padding, preserving full precision.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// SpatialRel defines the spatial relationship for geometry filters.
type SpatialRel string

// Supported spatial relationships for geometry filters.
const (
	SpatialRelIntersects         SpatialRel = "esriSpatialRelIntersects"
	SpatialRelContains           SpatialRel = "esriSpatialRelContains"
	SpatialRelWithin             SpatialRel = "esriSpatialRelWithin"
	SpatialRelTouches            SpatialRel = "esriSpatialRelTouches"
	SpatialRelOverlaps           SpatialRel = "esriSpatialRelOverlaps"
	SpatialRelEnvelopeIntersects SpatialRel = "esriSpatialRelEnvelopeIntersects"
)

// GeometryType for spatial filter inputs.
type GeometryType string

// Supported geometry types for spatial filter inputs.
const (
	GeometryTypeEnvelope GeometryType = "esriGeometryEnvelope"
	GeometryTypePoint    GeometryType = "esriGeometryPoint"
	GeometryTypePolygon  GeometryType = "esriGeometryPolygon"
)
