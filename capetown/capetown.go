// Package capetown provides named layer IDs and pre-built QueryParams
// for the City of Cape Town Open Data Portal.
//
// Base URL: https://citymaps.capetown.gov.za/agsext/rest/services/Theme_Based/Open_Data_Service/FeatureServer
//
// The layer IDs below are best-effort and may drift as the upstream service
// is republished. Use a [github.com/richardwooding/go-arcgis.Client] with
// ServiceInfo to confirm them against the live service.
package capetown

import (
	"fmt"

	arcgis "github.com/richardwooding/go-arcgis"
)

// BaseURL is the City of Cape Town Open Data Feature Service endpoint.
const BaseURL = "https://citymaps.capetown.gov.za/agsext/rest/services/Theme_Based/Open_Data_Service/FeatureServer"

// Layer IDs for well-known CCT datasets.
const (
	LayerLoadSheddingBlocks = 111
	LayerServiceRequests    = 1 // placeholder — verify against live service
	LayerWards              = 64
	LayerLandParcels        = 56
	LayerTaxiRoutes         = 108
	LayerPublicLighting     = 74
	LayerWaterQuality       = 229
	LayerHeritageInventory  = 49
)

// fieldSuburb is the suburb attribute name shared by several datasets.
const fieldSuburb = "SUBURB"

// --- Load Shedding ---

// LoadSheddingBlocks returns all load shedding zone polygons.
func LoadSheddingBlocks() arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID: LayerLoadSheddingBlocks,
		Fields:  []string{"BLOCK_NAME", "STAGE", "SUBURB_NAME"},
	}
}

// LoadSheddingBlocksForStage returns zones for a specific load shedding stage (1–8).
func LoadSheddingBlocksForStage(stage int) arcgis.QueryParams {
	p := LoadSheddingBlocks()
	p.Where = fmt.Sprintf("STAGE = %d", stage)
	return p
}

// --- Service Requests ---

// ServiceRequests returns open service requests, most recent first.
func ServiceRequests() arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID:       LayerServiceRequests,
		Fields:        []string{"SR_NUMBER", "DESCRIPTION", "STATUS", fieldSuburb, "CREATED_DATE"},
		OrderByFields: []string{"CREATED_DATE DESC"},
	}
}

// ServiceRequestsBySuburb filters service requests by suburb name.
func ServiceRequestsBySuburb(suburb string) arcgis.QueryParams {
	p := ServiceRequests()
	p.Where = fmt.Sprintf("%s = '%s'", fieldSuburb, suburb)
	return p
}

// --- Wards ---

// Wards returns all municipal ward boundaries.
func Wards() arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID: LayerWards,
		Fields:  []string{"WARD_ID", "WARD_NO", "COUNCILLOR", "YEAR"},
	}
}

// --- Land Parcels ---

// LandParcels returns cadastral land parcel polygons.
func LandParcels() arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID: LayerLandParcels,
		Fields:  []string{"ERF_NO", "LEGAL_STATUS", fieldSuburb, "AREA_SQM"},
	}
}

// --- Transport ---

// TaxiRoutes returns all registered taxi routes.
func TaxiRoutes() arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID: LayerTaxiRoutes,
		Fields:  []string{"ROUTE_NO", "FROM_RANK", "TO_RANK", "OPERATOR"},
	}
}

// --- Water ---

// WaterQualityResults returns inland water quality measurements, most recent first.
func WaterQualityResults() arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID:       LayerWaterQuality,
		Fields:        []string{"SITE_NAME", "SAMPLE_DATE", "PARAMETER", "RESULT", "UNIT"},
		OrderByFields: []string{"SAMPLE_DATE DESC"},
	}
}
