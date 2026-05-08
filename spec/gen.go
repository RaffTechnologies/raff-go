// Package spec contains types and an HTTP client generated from the public
// OpenAPI specification at docs/api-reference/openapi.yaml.
//
// Do not edit spec.gen.go by hand — run `make generate` from the raff-go
// repo root to regenerate it after changing the spec.
package spec

//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../docs/api-reference/openapi.yaml
