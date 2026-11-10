package arcgis_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

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

func TestItemInfo(t *testing.T) {
	var path string
	srv, last := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte(`{"id": "` + testItemID + `", "title": "Population Estimates 2025 (Suburb Level)",
			"snippet": "Estimated annual population. ", "description": "<p>Estimated</p>", "type": "Feature Service",
			"url": "https://services6.arcgis.com/x/arcgis/rest/services/Population_Estimates_2026/FeatureServer",
			"tags": ["population", "suburb"], "licenseInfo": "<p>Terms</p>", "modified": 1794298382000, "created": 1794232522000}`))
	})

	item, err := arcgis.ItemInfo(context.Background(), srv.URL+"/", testItemID)
	if err != nil {
		t.Fatalf("ItemInfo: %v", err)
	}
	if want := "/sharing/rest/content/items/" + testItemID; path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if last.Get("f") != "json" {
		t.Errorf("f = %q, want json", last.Get("f"))
	}
	if item.ID != testItemID || item.Title != "Population Estimates 2025 (Suburb Level)" || item.Type != "Feature Service" {
		t.Errorf("item = %+v", item)
	}
	if !strings.HasSuffix(item.URL, "/FeatureServer") || item.Snippet == "" || item.Description == "" || item.LicenseInfo == "" {
		t.Errorf("item = %+v, want url, snippet, description and licenseInfo", item)
	}
	if len(item.Tags) != 2 || item.Tags[0] != "population" {
		t.Errorf("tags = %v, want [population suburb]", item.Tags)
	}
	if got := item.ModifiedTime().UTC().Format(time.DateOnly); got != "2026-11-10" {
		t.Errorf("modified = %s, want 2026-11-10", got)
	}
	if item.CreatedTime().After(item.ModifiedTime()) {
		t.Errorf("created %v after modified %v", item.CreatedTime(), item.ModifiedTime())
	}
}

func TestItemInfoFillsMissingID(t *testing.T) {
	srv, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"title": "Web Map", "type": "Web Map"}`))
	})
	item, err := arcgis.ItemInfo(context.Background(), srv.URL, testItemID)
	if err != nil {
		t.Fatalf("ItemInfo: %v", err)
	}
	if item.ID != testItemID || item.URL != "" {
		t.Errorf("item = %+v, want the requested ID and no URL", item)
	}
	if !item.ModifiedTime().IsZero() {
		t.Errorf("modified = %v, want zero when unset", item.ModifiedTime())
	}
}

func TestItemInfoRejectsMalformedID(t *testing.T) {
	if _, err := arcgis.ItemInfo(context.Background(), "http://unused.invalid", "nope"); err == nil {
		t.Error("ItemInfo(nope) succeeded, want error")
	}
}
