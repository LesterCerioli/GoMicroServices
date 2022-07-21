package selector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type mockConn struct {
	attributes *attributes.Attributes
}

func (c mockConn) Address() resolver.Address {
	return resolver.Address{
		BalancerAttributes: c.attributes,
	}
}

func TestDefaultSelector_Select(t *testing.T) {

	t.Run("server is dyed", func(t *testing.T) {
		selector := defaultSelector{}

		conns := []Conn{
			mockConn{attributes: attributes.New("colors", NewColors("v1", "v2"))},
			mockConn{},
		}
		selectedConns := selector.Select(conns, balancer.PickInfo{
			FullMethodName: "foo",
			Ctx:            NewColorsContext(context.Background(), "v1"),
		})
		assert.Len(t, selectedConns, 1)
	})

	t.Run("server is not dyed", func(t *testing.T) {
		selector := defaultSelector{}

		conns := []Conn{
			mockConn{},
			mockConn{attributes: attributes.New("colors", NewColors())},
		}
		selectedConns := selector.Select(conns, balancer.PickInfo{
			FullMethodName: "foo",
			Ctx:            NewColorsContext(context.Background(), "v1"),
		})
		assert.Len(t, selectedConns, 2)
	})

}
func TestDefaultSelector_Name(t *testing.T) {
	selector := defaultSelector{}
	assert.Equal(t, "defaultSelector", selector.Name())
}
