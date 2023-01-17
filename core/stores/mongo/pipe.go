package mongo

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/zeromicro/go-zero/core/breaker"
)

type (
	// Pipe interface represents a mongo pipe.
	Pipe interface {
		All(result any) error
		AllowDiskUse() Pipe
		Batch(n int) Pipe
		Collation(collation *mgo.Collation) Pipe
		Explain(result any) error
		Iter() Iter
		One(result any) error
		SetMaxTime(d time.Duration) Pipe
	}

	promisedPipe struct {
		*mgo.Pipe
		promise keepablePromise
	}

	rejectedPipe struct{}
)

func (p promisedPipe) All(result any) error {
	return p.promise.keep(p.Pipe.All(result))
}

func (p promisedPipe) AllowDiskUse() Pipe {
	p.Pipe.AllowDiskUse()
	return p
}

func (p promisedPipe) Batch(n int) Pipe {
	p.Pipe.Batch(n)
	return p
}

func (p promisedPipe) Collation(collation *mgo.Collation) Pipe {
	p.Pipe.Collation(collation)
	return p
}

func (p promisedPipe) Explain(result any) error {
	return p.promise.keep(p.Pipe.Explain(result))
}

func (p promisedPipe) Iter() Iter {
	return promisedIter{
		Iter:    p.Pipe.Iter(),
		promise: p.promise,
	}
}

func (p promisedPipe) One(result any) error {
	return p.promise.keep(p.Pipe.One(result))
}

func (p promisedPipe) SetMaxTime(d time.Duration) Pipe {
	p.Pipe.SetMaxTime(d)
	return p
}

func (p rejectedPipe) All(result any) error {
	return breaker.ErrServiceUnavailable
}

func (p rejectedPipe) AllowDiskUse() Pipe {
	return p
}

func (p rejectedPipe) Batch(n int) Pipe {
	return p
}

func (p rejectedPipe) Collation(collation *mgo.Collation) Pipe {
	return p
}

func (p rejectedPipe) Explain(result any) error {
	return breaker.ErrServiceUnavailable
}

func (p rejectedPipe) Iter() Iter {
	return rejectedIter{}
}

func (p rejectedPipe) One(result any) error {
	return breaker.ErrServiceUnavailable
}

func (p rejectedPipe) SetMaxTime(d time.Duration) Pipe {
	return p
}