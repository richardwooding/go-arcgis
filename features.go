package arcgis

import "encoding/json"

// Feature represents a single ArcGIS feature with geometry and attributes.
type Feature struct {
	Geometry   json.RawMessage        `json:"geometry"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"` // GeoJSON alias
}

// FeatureSet is a collection of features returned from a query.
type FeatureSet struct {
	Features             []Feature `json:"features"`
	ExceededTransferLimit bool     `json:"exceededTransferLimit"`
	ObjectIDFieldName    string    `json:"objectIdFieldName,omitempty"`
	Fields               []Field   `json:"fields,omitempty"`
}

// Field describes a single attribute field in a layer.
type Field struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Alias string `json:"alias"`
}

// LayerInfo contains metadata about a feature layer.
type LayerInfo struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Description       string  `json:"description"`
	MaxRecordCount    int     `json:"maxRecordCount"`
	Fields            []Field `json:"fields"`
	GeometryType      string  `json:"geometryType"`
	SupportsStatistics bool   `json:"supportsStatistics"`
	SupportsPagination bool   `json:"supportsPagination"`
}

// ServiceInfo contains metadata about a feature service.
type ServiceInfo struct {
	ServiceDescription string      `json:"serviceDescription"`
	Layers             []LayerRef  `json:"layers"`
	Tables             []LayerRef  `json:"tables"`
}

// LayerRef is a lightweight layer reference in a service listing.
type LayerRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
