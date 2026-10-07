package arcgis_test

import (
	"context"
	"net/http"
	"testing"

	arcgis "github.com/richardwooding/go-arcgis"
)

func TestSimplifyParams(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})
	if _, err := arcgis.NewClient(srv.URL).Layer(2).Query().Simplify(0.00005, 6).First(context.Background()); err != nil {
		t.Fatalf("query: %v", err)
	}
	if got := last.Get("maxAllowableOffset"); got != "0.00005" {
		t.Errorf("maxAllowableOffset = %q, want 0.00005", got)
	}
	if got := last.Get("geometryPrecision"); got != "6" {
		t.Errorf("geometryPrecision = %q, want 6", got)
	}
}

func TestNoSimplifyByDefault(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})
	if _, err := arcgis.NewClient(srv.URL).Query(context.Background(), arcgis.QueryParams{LayerID: 2}); err != nil {
		t.Fatalf("query: %v", err)
	}
	if last.Has("maxAllowableOffset") || last.Has("geometryPrecision") {
		t.Errorf("unexpected simplify params: %v", *last)
	}
}
