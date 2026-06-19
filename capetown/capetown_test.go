package capetown_test

import (
	"strings"
	"testing"

	"github.com/richardwooding/go-arcgis/capetown"
)

func TestBaseURL(t *testing.T) {
	if !strings.HasSuffix(capetown.BaseURL, "/FeatureServer") {
		t.Errorf("BaseURL should point at a FeatureServer, got %q", capetown.BaseURL)
	}
}

func TestLoadSheddingBlocks(t *testing.T) {
	p := capetown.LoadSheddingBlocks()
	if p.LayerID != capetown.LayerLoadSheddingBlocks {
		t.Errorf("LayerID = %d, want %d", p.LayerID, capetown.LayerLoadSheddingBlocks)
	}
	if len(p.Fields) == 0 {
		t.Error("expected default fields to be set")
	}
}

func TestLoadSheddingBlocksForStage(t *testing.T) {
	p := capetown.LoadSheddingBlocksForStage(4)
	if p.Where != "STAGE = 4" {
		t.Errorf("Where = %q, want STAGE = 4", p.Where)
	}
	if p.LayerID != capetown.LayerLoadSheddingBlocks {
		t.Errorf("LayerID = %d", p.LayerID)
	}
}

func TestServiceRequestsBySuburb(t *testing.T) {
	p := capetown.ServiceRequestsBySuburb("Woodstock")
	if !strings.Contains(p.Where, "Woodstock") {
		t.Errorf("Where = %q, want it to reference Woodstock", p.Where)
	}
	if len(p.OrderByFields) == 0 {
		t.Error("expected service requests to be ordered")
	}
}

func TestNamedQueriesHaveLayerIDs(t *testing.T) {
	cases := map[string]int{
		"Wards":               capetown.Wards().LayerID,
		"LandParcels":         capetown.LandParcels().LayerID,
		"TaxiRoutes":          capetown.TaxiRoutes().LayerID,
		"WaterQualityResults": capetown.WaterQualityResults().LayerID,
		"ServiceRequests":     capetown.ServiceRequests().LayerID,
	}
	for name, id := range cases {
		if id <= 0 {
			t.Errorf("%s: layer ID = %d, want positive", name, id)
		}
	}
}
