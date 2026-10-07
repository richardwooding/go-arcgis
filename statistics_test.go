package arcgis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	arcgis "github.com/richardwooding/go-arcgis"
)

func TestStatisticsQuery(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": [
			{"attributes": {"Ward": "062", "n": 965}},
			{"attributes": {"Ward": "001", "n": 12}}
		]}`))
	})

	fs, err := arcgis.NewClient(srv.URL).Layer(0).Query().
		Where("Ward <> '#'").
		GroupBy("Ward").
		Statistics(arcgis.Statistic{Type: arcgis.StatCount, OnField: "ObjectId", OutName: "n"}).
		OrderBy("n DESC").
		First(context.Background())
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if got := last.Get("f"); got != "json" {
		t.Errorf("f = %q, want json", got)
	}
	if last.Has("outFields") {
		t.Errorf("outFields = %q, want unset for a statistics query", last.Get("outFields"))
	}
	if got := last.Get("groupByFieldsForStatistics"); got != "Ward" {
		t.Errorf("groupByFieldsForStatistics = %q, want Ward", got)
	}
	var stats []map[string]string
	if err := json.Unmarshal([]byte(last.Get("outStatistics")), &stats); err != nil {
		t.Fatalf("outStatistics not JSON: %v (%q)", err, last.Get("outStatistics"))
	}
	want := map[string]string{"statisticType": "count", "onStatisticField": "ObjectId", "outStatisticFieldName": "n"}
	if len(stats) != 1 || stats[0]["statisticType"] != want["statisticType"] ||
		stats[0]["onStatisticField"] != want["onStatisticField"] ||
		stats[0]["outStatisticFieldName"] != want["outStatisticFieldName"] {
		t.Errorf("outStatistics = %v, want [%v]", stats, want)
	}
	if len(fs.Features) != 2 || fs.Features[0].Attrs()["n"] != float64(965) {
		t.Errorf("features = %+v, want two groups led by n=965", fs.Features)
	}
}

func TestStatisticsOverridesGeoJSONFormat(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features": []}`))
	})
	_, err := arcgis.NewClient(srv.URL).Query(context.Background(), arcgis.QueryParams{
		Format:        arcgis.FormatGeoJSON,
		OutStatistics: []arcgis.Statistic{{Type: arcgis.StatSum, OnField: "VALUE", OutName: "total"}},
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got := last.Get("f"); got != "json" {
		t.Errorf("f = %q, want json", got)
	}
}

func TestUngroupedStatisticsOmitPaging(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features":[{"attributes":{"latest":1}}]}`))
	})
	_, err := arcgis.NewClient(srv.URL).Query(context.Background(), arcgis.QueryParams{
		OutStatistics: []arcgis.Statistic{{Type: arcgis.StatMax, OnField: "D", OutName: "latest"}},
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if last.Has("resultOffset") || last.Has("resultRecordCount") {
		t.Errorf("ungrouped statistics must not send paging params: %v", *last)
	}
}

func TestGroupedStatisticsKeepPaging(t *testing.T) {
	srv, last := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"features":[]}`))
	})
	_, err := arcgis.NewClient(srv.URL).Query(context.Background(), arcgis.QueryParams{
		GroupByFields: []string{"Ward"},
		OutStatistics: []arcgis.Statistic{{Type: arcgis.StatCount, OnField: "ID", OutName: "n"}},
		PageSize:      25,
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if last.Get("resultRecordCount") != "25" {
		t.Errorf("grouped statistics should keep resultRecordCount, got %q", last.Get("resultRecordCount"))
	}
}
