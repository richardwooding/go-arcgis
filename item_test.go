package arcgis_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	arcgis "github.com/richardwooding/go-arcgis"
)

const testItemID = "90711027d33940b5a06e96ad8a8f7ede"

func TestItemURL(t *testing.T) {
	var path string
	srv, last := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte(`{"id": "` + testItemID + `", "url": "https://services6.arcgis.com/x/arcgis/rest/services/SR/FeatureServer"}`))
	})

	got, err := arcgis.ItemURL(context.Background(), srv.URL+"/", testItemID)
	if err != nil {
		t.Fatalf("ItemURL: %v", err)
	}
	if want := "https://services6.arcgis.com/x/arcgis/rest/services/SR/FeatureServer"; got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
	if want := "/sharing/rest/content/items/" + testItemID; path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if last.Get("f") != "json" {
		t.Errorf("f = %q, want json", last.Get("f"))
	}
}

func TestItemURLRejectsMalformedID(t *testing.T) {
	for _, id := range []string{"", "../../etc", strings.ToUpper(testItemID), testItemID + "0"} {
		if _, err := arcgis.ItemURL(context.Background(), "http://unused.invalid", id); err == nil {
			t.Errorf("ItemURL(%q) succeeded, want error", id)
		}
	}
}

func TestItemURLWithoutServiceURL(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id": "` + testItemID + `", "type": "Web Map"}`))
	})
	if _, err := arcgis.ItemURL(context.Background(), srv.URL, testItemID); err == nil {
		t.Error("want error for an item with no url")
	}
}

func TestItemURLSurfacesPortalError(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error": {"code": 400, "message": "Item does not exist or is inaccessible."}}`))
	})
	_, err := arcgis.ItemURL(context.Background(), srv.URL, testItemID)
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("err = %v, want the portal's message", err)
	}
}
