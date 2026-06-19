package arcgis_test

import (
	"context"
	"fmt"
	"log"

	arcgis "github.com/richardwooding/go-arcgis"
	"github.com/richardwooding/go-arcgis/capetown"
)

func Example_structStyle() {
	client := arcgis.NewClient(capetown.BaseURL)
	ctx := context.Background()

	// Struct-style — explicit, composable, serializable
	features, err := client.Query(ctx, arcgis.QueryParams{
		LayerID:  capetown.LayerLoadSheddingBlocks,
		Where:    "STAGE = 4",
		Fields:   []string{"BLOCK_NAME", "STAGE"},
		PageSize: 100,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(features.Features))
}

func Example_fluentStyle() {
	client := arcgis.NewClient(capetown.BaseURL)
	ctx := context.Background()

	// Fluent style — readable, chainable
	features, err := client.Layer(capetown.LayerLoadSheddingBlocks).
		Query().
		Where("STAGE = 4").
		Fields("BLOCK_NAME", "STAGE").
		WithinEnvelope(18.4, -34.0, 18.6, -33.8).
		PageSize(100).
		All(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(features))
}

func Example_namedQuery() {
	client := arcgis.NewClient(capetown.BaseURL)
	ctx := context.Background()

	// Named CCT query — pre-built params, optionally extended
	features, err := client.Layer(capetown.LayerLoadSheddingBlocks).
		Query().
		From(capetown.LoadSheddingBlocksForStage(4)).
		WithinEnvelope(18.4, -34.0, 18.6, -33.8). // further refined
		All(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(features))
}

func Example_paginatedAll() {
	client := arcgis.NewClient(capetown.BaseURL)
	ctx := context.Background()

	// QueryAll handles pagination transparently
	all, err := client.QueryAll(ctx, capetown.WaterQualityResults())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total water quality records: %d\n", len(all))
}

func Example_countOnly() {
	client := arcgis.NewClient(capetown.BaseURL)
	ctx := context.Background()

	// Just get a count — no feature data transferred
	count, err := client.Layer(capetown.LayerServiceRequests).
		Query().
		From(capetown.ServiceRequestsBySuburb("Woodstock")).
		Count(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Open requests in Woodstock: %d\n", count)
}

func Example_inspectParams() {
	client := arcgis.NewClient(capetown.BaseURL)

	// QueryParams is just a struct — inspect, log, or marshal it
	p := client.Layer(capetown.LayerWards).
		Query().
		Where("YEAR = 2021").
		WithoutGeometry().
		Params() // returns QueryParams without executing

	fmt.Printf("Layer: %d, Where: %s\n", p.LayerID, p.Where)
}
