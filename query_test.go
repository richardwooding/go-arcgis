package arcgis_test

import (
	"context"
	"net/http"
	"testing"

	arcgis "github.com/richardwooding/go-arcgis"
)

func TestDefaults_AppliedViaRequest(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL)
	if _, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 7}); err != nil {
		t.Fatalf("Query: %v", err)
	}

	want := map[string]string{
		"where":             "1=1",
		"f":                 "geojson",
		"resultRecordCount": "1000",
		"resultOffset":      "0",
		"outFields":         "*",
	}
	for k, v := range want {
		if got := last.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestEnvelopeFilter(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL)
	_, err := client.Layer(7).Query().
		WithinEnvelope(18.4, -34.0, 18.6, -33.8).
		First(context.Background())
	if err != nil {
		t.Fatalf("First: %v", err)
	}

	if got, want := last.Get("geometry"), "18.4,-34,18.6,-33.8"; got != want {
		t.Errorf("geometry = %q, want %q", got, want)
	}
	if got := last.Get("geometryType"); got != "esriGeometryEnvelope" {
		t.Errorf("geometryType = %q", got)
	}
	if got := last.Get("spatialRel"); got != "esriSpatialRelIntersects" {
		t.Errorf("spatialRel = %q", got)
	}
}

func TestPointFilter(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL)
	_, err := client.Layer(7).Query().
		IntersectsPoint(18.42, -33.92).
		SpatialRel(arcgis.SpatialRelWithin).
		First(context.Background())
	if err != nil {
		t.Fatalf("First: %v", err)
	}

	if got, want := last.Get("geometry"), "18.42,-33.92"; got != want {
		t.Errorf("geometry = %q, want %q", got, want)
	}
	if got := last.Get("geometryType"); got != "esriGeometryPoint" {
		t.Errorf("geometryType = %q", got)
	}
	if got := last.Get("spatialRel"); got != "esriSpatialRelWithin" {
		t.Errorf("spatialRel = %q", got)
	}
}

func TestBuilder_FieldsOrderGroupAndGeometry(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL)
	_, err := client.Layer(7).Query().
		Where("STATUS = 'OPEN'").
		Fields("A", "B").
		OrderBy("CREATED DESC").
		GroupBy("SUBURB").
		WithoutGeometry().
		PageSize(50).
		Offset(100).
		All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}

	checks := map[string]string{
		"where":                      "STATUS = 'OPEN'",
		"outFields":                  "A,B",
		"orderByFields":              "CREATED DESC",
		"groupByFieldsForStatistics": "SUBURB",
		"returnGeometry":             "false",
		"resultRecordCount":          "50",
		"resultOffset":               "100",
	}
	for k, v := range checks {
		if got := last.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestWhereClauseIsEscaped(t *testing.T) {
	// A WHERE clause with spaces, quotes and an ampersand must round-trip
	// intact through URL encoding.
	const where = "NAME = 'A & B' AND X > 1"
	srv, last := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		// RawQuery must not contain an unescaped space.
		if got := r.URL.RawQuery; containsRune(got, ' ') {
			t.Errorf("raw query contains unescaped space: %q", got)
		}
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL)
	if _, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 7, Where: where}); err != nil {
		t.Fatalf("Query: %v", err)
	}
	if got := last.Get("where"); got != where {
		t.Errorf("decoded where = %q, want %q", got, where)
	}
}

func TestParams_Inspect(t *testing.T) {
	client := arcgis.NewClient("http://example.test")
	p := client.Layer(64).Query().
		Where("YEAR = 2021").
		WithoutGeometry().
		Params()

	if p.LayerID != 64 {
		t.Errorf("LayerID = %d, want 64", p.LayerID)
	}
	if p.Where != "YEAR = 2021" {
		t.Errorf("Where = %q", p.Where)
	}
	if p.ReturnGeometry == nil || *p.ReturnGeometry {
		t.Errorf("ReturnGeometry should be set to false")
	}
}

func TestFrom_PreservesLayerID(t *testing.T) {
	base := arcgis.QueryParams{LayerID: 999, Where: "X = 1", Fields: []string{"A"}}
	client := arcgis.NewClient("http://example.test")
	p := client.Layer(7).Query().From(base).Params()

	if p.LayerID != 7 {
		t.Errorf("LayerID = %d, want 7 (preserved over From)", p.LayerID)
	}
	if p.Where != "X = 1" || len(p.Fields) != 1 {
		t.Errorf("From did not merge base params: %+v", p)
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
