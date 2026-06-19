package arcgis

// Envelope is a bounding box spatial filter.
type Envelope struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// Point is a single coordinate.
type Point struct {
	X float64
	Y float64
}

// SpatialRel defines the spatial relationship for geometry filters.
type SpatialRel string

const (
	SpatialRelIntersects     SpatialRel = "esriSpatialRelIntersects"
	SpatialRelContains       SpatialRel = "esriSpatialRelContains"
	SpatialRelWithin         SpatialRel = "esriSpatialRelWithin"
	SpatialRelTouches        SpatialRel = "esriSpatialRelTouches"
	SpatialRelOverlaps       SpatialRel = "esriSpatialRelOverlaps"
	SpatialRelEnvelopeIntersects SpatialRel = "esriSpatialRelEnvelopeIntersects"
)

// GeometryType for spatial filter inputs.
type GeometryType string

const (
	GeometryTypeEnvelope GeometryType = "esriGeometryEnvelope"
	GeometryTypePoint    GeometryType = "esriGeometryPoint"
	GeometryTypePolygon  GeometryType = "esriGeometryPolygon"
)
