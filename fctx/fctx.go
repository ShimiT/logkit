package fctx

// Package reqctx provides functionality for creating context keys unique to a
// given service. Package reqctx should be initialized by calling Init() at the
// startup of any service using it. If you find svc_unset in metrics Init() is
// most likely not being called

import (
	"context"
	"sync"
)

// ReqKey represents a component of a context value tied specifically to a
// service and endpoint.
type ReqKey int

// ReqKey constants represent the different values that may be placed into the
// context, to allowed called functions to emit metrics that can be associated
// with a specific service and endpoint combination.
const (
	ReqUnset ReqKey = iota
	ReqEndpoint
	ReqAuth
)

// SvcKey represents a service.
type SvcKey int

// SvcKey constants represent the different services, in order to allow
// creation of service-specific context keys.
const (
	SvcUnset SvcKey = iota
)

// CtxKey ties a SvcKey and ReqKey together, allowing unique context values per
// service without name collision, which would otherwise occur without defining
// what service the context value belongs to.
type CtxKey struct {
	SvcKey SvcKey
	ReqKey ReqKey
}

var (
	svc  SvcKey
	once sync.Once
)

var serviceNameToSvcKey = map[string]SvcKey{}

// Init initializes the name of the service that will be used when
// constructing service-specific context keys, which is done to prevent
// inter-service name collisions for context values.
func Init(service string) {
	if svcKey, ok := serviceNameToSvcKey[service]; ok {
		once.Do(func() { svc = svcKey })
		return
	}
	once.Do(func() { svc = SvcUnset })
}

// NewContextKey creates a key suitable for use as a key in context.WithValue
// that is unique to the service that created the key.
func NewContextKey(reqKey ReqKey) CtxKey {
	return CtxKey{svc, reqKey}
}

// selectedTags determines what tags will be extracted from the context in the
// tagsFromContext function.
var selectedTags = map[string]ReqKey{
	"endpoint": ReqEndpoint,
}

// MetricsTagsFromContext extracts pre-defined tags from a context, suitable
// for passing to the metrics With() tag-defining function.
func MetricsTagsFromContext(ctx context.Context) []string {
	tags := make([]string, 0, len(selectedTags)*2)
	for tagname, tagkey := range selectedTags {
		tags = append(tags, tagname)
		if svc == SvcUnset {
			tags = append(tags, "svc_unset")
			continue
		}
		val := ctx.Value(NewContextKey(tagkey))
		// Value was not present in the context.
		if val == nil {
			tags = append(tags, "unset")
			continue
		}
		tags = append(tags, val.(string))
	}
	return tags
}

// LogTagsFromContext extracts pre-defined tags from a a context, suitable
// for passing to the logging With() context-defining function.
func LogTagsFromContext(ctx context.Context) []interface{} {
	tags := MetricsTagsFromContext(ctx)
	intfTags := make([]interface{}, len(tags))
	for i, _ := range tags {
		intfTags[i] = tags[i]
	}
	return intfTags
}
