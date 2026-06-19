package arcgis_test

import (
	"context"
	"fmt"
	"log"

	arcgis "github.com/richardwooding/go-arcgis"
)

// baseURL is an example FeatureServer root. Replace it with the service you
// are querying.
const baseURL = "https://example.gov/arcgis/rest/services/Theme/Service/FeatureServer"

// loadSheddingForStage is the kind of named, pre-built query an application
// or companion package might expose.
func loadSheddingForStage(stage int) arcgis.QueryParams {
	return arcgis.QueryParams{
		LayerID: 111,
		Where:   fmt.Sprintf("STAGE = %d", stage),
		Fields:  []string{"BLOCK_NAME", "STAGE"},
	}
}

func Example_structStyle() {
	client := arcgis.NewClient(baseURL)
	ctx := context.Background()

	// Struct-style — explicit, composable, serializable
	features, err := client.Query(ctx, arcgis.QueryParams{
		LayerID:  111,
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
	client := arcgis.NewClient(baseURL)
	ctx := context.Background()

	// Fluent style — readable, chainable
	features, err := client.Layer(111).
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
	client := arcgis.NewClient(baseURL)
	ctx := context.Background()

	// Named query — pre-built params, optionally extended
	features, err := client.Layer(111).
		Query().
		From(loadSheddingForStage(4)).
		WithinEnvelope(18.4, -34.0, 18.6, -33.8). // further refined
		All(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(features))
}

func Example_paginatedAll() {
	client := arcgis.NewClient(baseURL)
	ctx := context.Background()

	// QueryAll handles pagination transparently
	all, err := client.QueryAll(ctx, arcgis.QueryParams{LayerID: 229})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total records: %d\n", len(all))
}

func Example_countOnly() {
	client := arcgis.NewClient(baseURL)
	ctx := context.Background()

	// Just get a count — no feature data transferred
	count, err := client.Layer(1).
		Query().
		Where("SUBURB = 'Woodstock'").
		Count(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Matching records: %d\n", count)
}

func Example_inspectParams() {
	client := arcgis.NewClient(baseURL)

	// QueryParams is just a struct — inspect, log, or marshal it
	p := client.Layer(64).
		Query().
		Where("YEAR = 2021").
		WithoutGeometry().
		Params() // returns QueryParams without executing

	fmt.Printf("Layer: %d, Where: %s\n", p.LayerID, p.Where)
	// Output: Layer: 64, Where: YEAR = 2021
}
