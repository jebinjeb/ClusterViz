// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.13.0 DO NOT EDIT.
package api

// Cluster defines model for Cluster.
type Cluster struct {
	Id   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}
