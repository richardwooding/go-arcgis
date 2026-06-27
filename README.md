# go-arcgis

[![Go Reference](https://pkg.go.dev/badge/github.com/richardwooding/go-arcgis.svg)](https://pkg.go.dev/github.com/richardwooding/go-arcgis)
[![Go](https://github.com/richardwooding/go-arcgis/actions/workflows/go.yml/badge.svg)](https://github.com/richardwooding/go-arcgis/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A small, dependency-free Go client for querying **ArcGIS Feature Services**
over their REST API. It handles pagination, counts, and object-ID-only
queries, and offers two interchangeable styles: a plain struct API and a fluent
builder.

```go
client := arcgis.NewClient(baseURL)

features, err := client.Layer(7).Query().
    Where("STAGE = 4").
    Fields("BLOCK_NAME", "STAGE").
    All(ctx) // paginates automatically
```

## Install

```sh
go get github.com/richardwooding/go-arcgis
```

Requires Go 1.26+. No third-party dependencies.

## Two styles, one engine

The fluent builder and the struct API produce the same `QueryParams` and hit
the same code path — pick whichever reads better at the call site.

### Struct style — explicit and serializable

```go
fs, err := client.Query(ctx, arcgis.QueryParams{
    LayerID:  7,
    Where:    "STAGE = 4",
    Fields:   []string{"BLOCK_NAME", "STAGE"},
    PageSize: 100,
})
```

### Fluent style — readable and chainable

```go
features, err := client.Layer(7).Query().
    Where("STAGE = 4").
    Fields("BLOCK_NAME", "STAGE").
    WithinEnvelope(18.4, -34.0, 18.6, -33.8).
    All(ctx)
```

## Querying

| Call | Returns | Notes |
| --- | --- | --- |
| `Query` / `.First` | `*FeatureSet` | a single page |
| `QueryAll` / `.All` | `[]Feature` | follows `exceededTransferLimit` until exhausted |
| `QueryCount` / `.Count` | `int` | no feature data transferred |
| `QueryIDs` / `.IDs` | `[]int64` | object IDs only |

Pagination is automatic: `QueryAll` keeps advancing `resultOffset` while the
service reports `exceededTransferLimit`.

Attributes are exposed uniformly regardless of the response format —
`Feature.Attrs()` returns the Esri-JSON `attributes` map, falling back to the
GeoJSON `properties` map:

```go
for _, f := range features {
    fmt.Println(f.Attrs()["BLOCK_NAME"])
}
```

## Spatial filters

```go
// Bounding box (lon/lat)
client.Layer(7).Query().WithinEnvelope(18.4, -34.0, 18.6, -33.8)

// Point with an explicit relationship
client.Layer(7).Query().
    IntersectsPoint(18.42, -33.92).
    SpatialRel(arcgis.SpatialRelWithin)
```

## Output format

GeoJSON is the default. Switch to Esri JSON when you need its richer metadata
(`objectIdFieldName`, field definitions):

```go
client.Layer(7).Query().Format(arcgis.FormatJSON).First(ctx)
```

`Count` and `IDs` always use Esri JSON internally, since that is the only
format ArcGIS returns those responses in.

## Options

```go
arcgis.NewClient(baseURL,
    arcgis.WithToken("…"),                 // authenticated services
    arcgis.WithTimeout(10*time.Second),    // or:
    arcgis.WithHTTPClient(myClient),       // bring your own transport
)
```

A token, when set, is appended to every request — queries, counts, service and
layer metadata.

## Errors

ArcGIS frequently reports failures with an HTTP 200 status and an error
envelope in the body. `go-arcgis` surfaces these as `*arcgis.APIError`:

```go
_, err := client.Query(ctx, p)
var apiErr *arcgis.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.Code, apiErr.Message)
}
```

## Named services

Build named layer IDs and pre-built `QueryParams` for a specific service, then
extend them at the call site with `.From(...)`:

```go
func loadSheddingForStage(stage int) arcgis.QueryParams {
    return arcgis.QueryParams{
        LayerID: 111,
        Where:   fmt.Sprintf("STAGE = %d", stage),
        Fields:  []string{"BLOCK_NAME", "STAGE"},
    }
}

features, err := client.Layer(111).Query().
    From(loadSheddingForStage(4)).
    WithinEnvelope(18.4, -34.0, 18.6, -33.8).
    All(ctx)
```

For the City of Cape Town Open Data Portal, the companion package
[`capetown-opendata`](https://github.com/richardwooding/capetown-opendata)
ships these constructors and verified layer IDs ready to use.

## Ports

A PHP port that tracks this library's design is available:
[php-arcgis](https://github.com/richardwooding/php-arcgis)
(`composer require richardwooding/arcgis`).

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE)
