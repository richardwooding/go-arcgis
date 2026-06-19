package arcgis_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	arcgis "github.com/richardwooding/go-arcgis"
)

// newServer starts a test server whose handler captures the most recent
// request's query values for assertion.
func newServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *url.Values) {
	t.Helper()
	var last url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last = r.URL.Query()
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &last
}

func TestQuery_EsriJSON(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"objectIdFieldName": "OBJECTID",
			"exceededTransferLimit": false,
			"features": [
				{"geometry": {"x": 1, "y": 2}, "attributes": {"NAME": "Block A", "STAGE": 4}}
			]
		}`))
	})

	client := arcgis.NewClient(srv.URL)
	fs, err := client.Query(context.Background(), arcgis.QueryParams{
		LayerID: 7,
		Where:   "STAGE = 4",
		Format:  arcgis.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	if last.Get("where") != "STAGE = 4" {
		t.Errorf("where = %q, want %q", last.Get("where"), "STAGE = 4")
	}
	if got := last.Get("f"); got != "json" {
		t.Errorf("f = %q, want json", got)
	}
	if len(fs.Features) != 1 {
		t.Fatalf("got %d features, want 1", len(fs.Features))
	}
	if got := fs.Features[0].Attrs()["NAME"]; got != "Block A" {
		t.Errorf("NAME = %v, want Block A", got)
	}
	if fs.ObjectIDFieldName != "OBJECTID" {
		t.Errorf("ObjectIDFieldName = %q", fs.ObjectIDFieldName)
	}
}

func TestQuery_GeoJSONProperties(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"type": "FeatureCollection",
			"features": [
				{"type": "Feature", "geometry": null, "properties": {"NAME": "Ward 1"}}
			]
		}`))
	})

	client := arcgis.NewClient(srv.URL)
	fs, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 1})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if got := fs.Features[0].Attrs()["NAME"]; got != "Ward 1" {
		t.Errorf("Attrs()[NAME] = %v, want Ward 1", got)
	}
}

func TestQueryAll_Paginates(t *testing.T) {
	var calls int
	srv, last := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		offset := r.URL.Query().Get("resultOffset")
		switch offset {
		case "0":
			_, _ = w.Write([]byte(`{"exceededTransferLimit": true, "features": [
				{"attributes": {"id": 1}}, {"attributes": {"id": 2}}]}`))
		default: // offset == "2"
			_, _ = w.Write([]byte(`{"exceededTransferLimit": false, "features": [
				{"attributes": {"id": 3}}]}`))
		}
	})

	client := arcgis.NewClient(srv.URL)
	all, err := client.QueryAll(context.Background(), arcgis.QueryParams{LayerID: 7})
	if err != nil {
		t.Fatalf("QueryAll: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d features across pages, want 3", len(all))
	}
	if calls != 2 {
		t.Errorf("server called %d times, want 2", calls)
	}
	if last.Get("resultOffset") != "2" {
		t.Errorf("final resultOffset = %q, want 2", last.Get("resultOffset"))
	}
}

func TestQueryCount(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"count": 42}`))
	})

	client := arcgis.NewClient(srv.URL)
	n, err := client.QueryCount(context.Background(), arcgis.QueryParams{LayerID: 7})
	if err != nil {
		t.Fatalf("QueryCount: %v", err)
	}
	if n != 42 {
		t.Errorf("count = %d, want 42", n)
	}
	if last.Get("returnCountOnly") != "true" {
		t.Errorf("returnCountOnly = %q, want true", last.Get("returnCountOnly"))
	}
	if last.Get("f") != "json" {
		t.Errorf("count forces f=json, got %q", last.Get("f"))
	}
}

func TestQueryIDs(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"objectIds": [10, 20, 30]}`))
	})

	client := arcgis.NewClient(srv.URL)
	ids, err := client.QueryIDs(context.Background(), arcgis.QueryParams{LayerID: 7})
	if err != nil {
		t.Fatalf("QueryIDs: %v", err)
	}
	if len(ids) != 3 || ids[0] != 10 || ids[2] != 30 {
		t.Errorf("ids = %v, want [10 20 30]", ids)
	}
	if last.Get("returnIdsOnly") != "true" {
		t.Errorf("returnIdsOnly = %q, want true", last.Get("returnIdsOnly"))
	}
}

func TestServiceInfo(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"serviceDescription": "Open Data",
			"layers": [{"id": 1, "name": "Wards"}],
			"tables": []}`))
	})

	client := arcgis.NewClient(srv.URL)
	info, err := client.ServiceInfo(context.Background())
	if err != nil {
		t.Fatalf("ServiceInfo: %v", err)
	}
	if info.ServiceDescription != "Open Data" {
		t.Errorf("ServiceDescription = %q", info.ServiceDescription)
	}
	if len(info.Layers) != 1 || info.Layers[0].Name != "Wards" {
		t.Errorf("Layers = %+v", info.Layers)
	}
}

func TestLayerInfo(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/7" {
			http.Error(w, "wrong path: "+r.URL.Path, http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"id": 7, "name": "Blocks", "geometryType": "esriGeometryPolygon",
			"maxRecordCount": 2000, "supportsPagination": true}`))
	})

	client := arcgis.NewClient(srv.URL)
	info, err := client.Layer(7).Info(context.Background())
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.Name != "Blocks" || info.MaxRecordCount != 2000 || !info.SupportsPagination {
		t.Errorf("LayerInfo = %+v", info)
	}
}

func TestAPIError_On200(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		// ArcGIS returns a 200 with an error envelope.
		_, _ = w.Write([]byte(`{"error": {"code": 400, "message": "Invalid query",
			"details": ["Unable to parse where clause"]}}`))
	})

	client := arcgis.NewClient(srv.URL)
	_, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 7})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *arcgis.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %v", err)
	}
	if apiErr.Code != 400 || apiErr.Message != "Invalid query" {
		t.Errorf("APIError = %+v", apiErr)
	}
}

func TestHTTPErrorStatus(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	client := arcgis.NewClient(srv.URL)
	_, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 7})
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}

func TestWithToken(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL, arcgis.WithToken("secret-token"))
	if _, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 7}); err != nil {
		t.Fatalf("Query: %v", err)
	}
	if last.Get("token") != "secret-token" {
		t.Errorf("token = %q, want secret-token", last.Get("token"))
	}
}

func TestTokenAppliedToServiceInfo(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"serviceDescription": "x"}`))
	})

	client := arcgis.NewClient(srv.URL, arcgis.WithToken("t"))
	if _, err := client.ServiceInfo(context.Background()); err != nil {
		t.Fatalf("ServiceInfo: %v", err)
	}
	if last.Get("token") != "t" {
		t.Errorf("ServiceInfo should carry token, got %q", last.Get("token"))
	}
}

func TestContextCancelled(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := arcgis.NewClient(srv.URL)
	_, err := client.Query(ctx, arcgis.QueryParams{LayerID: 7})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWithTimeout(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"features": []}`))
	})

	client := arcgis.NewClient(srv.URL, arcgis.WithTimeout(5*time.Millisecond))
	_, err := client.Query(context.Background(), arcgis.QueryParams{LayerID: 7})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
