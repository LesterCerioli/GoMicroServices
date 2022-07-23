package md

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type (
	// Metadata represents the metadata of the service.
	Metadata map[string][]string
	mdKey    struct{}
)

// Append appends a set of data.
func (m Metadata) Append(k string, values ...string) {
	k = strings.ToLower(k)
	m[k] = append(m[k], values...)
}

// Keys returns all keys.
func (m Metadata) Keys() []string {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, strings.ToLower(k))
	}

	return keys
}

// Set sets a set of data
func (m Metadata) Set(key string, values ...string) {
	m[strings.ToLower(key)] = values
}

// Get gets the first element of the value corresponding to the specified key.
func (m Metadata) Get(key string) string {
	values, ok := m[strings.ToLower(key)]
	if !ok {
		return ""
	}
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

// Values gets all elements with the specified key.
func (m Metadata) Values(key string) []string {
	return m[strings.ToLower(key)]
}

// Delete deletes all elements with the specified key.
func (m Metadata) Delete(key string) {
	delete(m, strings.ToLower(key))
}

// Carrier Implement the Carrier interface.
func (m Metadata) Carrier() (Metadata, error) {
	return m.Clone(), nil
}

// Clone clones a Metadata.
func (m Metadata) Clone() Metadata {
	md := make(Metadata, len(m))
	for k, v := range m {
		md[strings.ToLower(k)] = v
	}

	return md
}

// FromContext extracts Metadata from context.
func FromContext(ctx context.Context) Metadata {
	value := ctx.Value(mdKey{})
	if value == nil {
		return Metadata{}
	}

	return value.(Metadata)
}

// NewContext creates a new metadata context.
func NewContext(ctx context.Context, carrier Carrier) context.Context {
	md, err := carrier.Carrier()
	if err != nil {
		logx.WithContext(ctx).Error(err)
		return ctx
	}

	return context.WithValue(ctx, mdKey{}, md)
}

func ValuesFromContext(ctx context.Context, key string) []string {
	return FromContext(ctx).Values(key)
}
