package arcgis

import (
	"encoding/json"
	"fmt"
)

// Feature represents a single ArcGIS feature with geometry and attributes.
//
// The geometry is left as raw JSON because its shape depends on the requested
// format (Esri JSON vs GeoJSON) and the layer's geometry type. Attributes are
// populated for Esri JSON responses; Properties for GeoJSON.
type Feature struct {
	Geometry   json.RawMessage `json:"geometry"`
	Attributes map[string]any  `json:"attributes,omitempty"` // Esri JSON
	Properties map[string]any  `json:"properties,omitempty"` // GeoJSON
}

// Attrs returns the feature's attribute map regardless of response format,
// preferring Esri JSON attributes and falling back to GeoJSON properties.
func (f Feature) Attrs() map[string]any {
	if f.Attributes != nil {
		return f.Attributes
	}
	return f.Properties
}

// FeatureSet is a collection of features returned from a query.
type FeatureSet struct {
	Features              []Feature `json:"features"`
	ExceededTransferLimit bool      `json:"exceededTransferLimit"`
	ObjectIDFieldName     string    `json:"objectIdFieldName,omitempty"`
	Fields                []Field   `json:"fields,omitempty"`
}

// Field describes a single attribute field in a layer.
type Field struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Alias string `json:"alias"`
}

// LayerInfo contains metadata about a feature layer.
type LayerInfo struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Type               string  `json:"type"`
	Description        string  `json:"description"`
	MaxRecordCount     int     `json:"maxRecordCount"`
	Fields             []Field `json:"fields"`
	GeometryType       string  `json:"geometryType"`
	SupportsStatistics bool    `json:"supportsStatistics"`
	SupportsPagination bool    `json:"supportsPagination"`
}

// ServiceInfo contains metadata about a feature service.
type ServiceInfo struct {
	ServiceDescription string     `json:"serviceDescription"`
	Layers             []LayerRef `json:"layers"`
	Tables             []LayerRef `json:"tables"`
}

// LayerRef is a lightweight layer reference in a service listing.
type LayerRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// APIError is an error returned by an ArcGIS service. ArcGIS commonly reports
// failures with an HTTP 200 status and an error envelope in the body, so this
// is surfaced even on otherwise-successful responses.
type APIError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("arcgis error %d: %s (%v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("arcgis error %d: %s", e.Code, e.Message)
}

// errorEnvelope is the wrapper ArcGIS uses to report errors in a 200 response.
type errorEnvelope struct {
	Error *APIError `json:"error"`
}
