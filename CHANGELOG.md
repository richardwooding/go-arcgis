# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/richardwooding/go-arcgis/releases/tag/v0.1.0
