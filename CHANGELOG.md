# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-10-07

### Added
- `QueryParams.MaxAllowableOffset` and `QueryParams.GeometryPrecision`, sent as
  `maxAllowableOffset` and `geometryPrecision`, plus `QueryBuilder.Simplify`.
  The server generalises geometries before returning them, which can halve
  polygon payloads.

## [0.4.0] - 2026-10-05

### Added
- Server-side statistics: `QueryParams.OutStatistics` and
  `QueryBuilder.Statistics(...)`, with `Statistic` and the `StatCount`,
  `StatSum`, `StatAvg`, `StatMin`, `StatMax`, `StatStddev` and `StatVar`
  aggregates. Combine with `GroupBy`; statistics responses are always Esri JSON.
- `ItemURL(ctx, portalURL, itemID)` resolves a portal item ID to its current
  service URL, with `ArcGISOnline` as the public portal. Hosted datasets are
  often republished under new service names while keeping their item ID.

## [0.3.0] - 2026-09-08

### Changed
- Raised the minimum Go version to 1.27 and applied the Go 1.27 modernizers.
  No API changes.

### Added
- GitHub Pages documentation site, with OpenGraph/Twitter preview metadata.
- README links to the [php-arcgis](https://github.com/richardwooding/php-arcgis)
  port and to sponsorship.

## [0.2.1] - 2026-06-19

### Fixed
- Queries whose encoded parameters would exceed URL length limits are now sent
  as `POST`, avoiding the 404s the ArcGIS server returns for over-long GETs.

## [0.2.0] - 2026-06-19

### Added
- `QueryParams.InSR` and `QueryBuilder.InSR(wkid)` to declare the spatial
  reference of the input geometry.
- `Polygon` geometry and `QueryBuilder.WithinPolygon(rings)` for polygon
  spatial filters.

## [0.1.1] - 2026-06-19

### Added
- `QueryParams.ReturnDistinctValues` and `QueryBuilder.DistinctValues()`.

### Removed
- The `capetown` subpackage has moved to its own module,
  [`capetown-opendata`](https://github.com/richardwooding/capetown-opendata).
  Import `github.com/richardwooding/capetown-opendata` instead of
  `github.com/richardwooding/go-arcgis/capetown`.

## [0.1.0] - 2026-06-19

Initial release.

### Added
- `Client` with `NewClient(baseURL, ...ClientOption)`, configurable via
  `WithToken`, `WithTimeout`, and `WithHTTPClient`. A token, when set, is
  applied to every request.
- Struct-style querying: `Client.Query`, `QueryAll` (automatic pagination via
  `exceededTransferLimit`), `QueryCount`, and `QueryIDs`.
- Fluent `QueryBuilder` (`Client.Layer(id).Query()`) with `Where`, `Fields`,
  `WithinEnvelope`, `IntersectsPoint`, `SpatialRel`, `OrderBy`, `GroupBy`,
  `Offset`, `PageSize`, `WithoutGeometry`, `Format`, and `From`; terminal
  methods `First`, `All`, `Count`, and `IDs`.
- Service and layer metadata via `Client.ServiceInfo` and `Client.LayerInfo` /
  `LayerClient.Info`.
- `Feature.Attrs()` returning attributes regardless of GeoJSON vs Esri JSON
  format.
- `APIError` surfacing ArcGIS error envelopes returned with an HTTP 200 status.
- `capetown` subpackage with named layer IDs and pre-built `QueryParams` for the
  City of Cape Town Open Data Portal.

[Unreleased]: https://github.com/richardwooding/go-arcgis/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/richardwooding/go-arcgis/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/richardwooding/go-arcgis/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/richardwooding/go-arcgis/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/richardwooding/go-arcgis/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/richardwooding/go-arcgis/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/richardwooding/go-arcgis/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/richardwooding/go-arcgis/releases/tag/v0.1.0
