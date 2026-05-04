// Package syncx provides generic synchronization primitives and utilities.
package syncx

import (
	"reflect"
	"sync"
)

// Pool is a generic wrapper around sync.Pool for managing reusable objects of any type.
type Pool[T any] struct {
	pool *sync.Pool
}

// NewPool creates a new generic Pool that uses newFunc to allocate new items.
func NewPool[T any](newFunc func() *T) *Pool[T] {
	tType := reflect.TypeOf((*T)(nil)).Elem()
	if tType.Kind() == reflect.Ptr {
		panic("NewPool requires a non-pointer type T")
	}
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get retrieves an item from the Pool, allocating a new one if necessary.
func (p *Pool[T]) Get() *T {
	return p.pool.Get().(*T)
}

// Put returns an item to the Pool for future reuse.
func (p *Pool[T]) Put(item *T) {
	p.pool.Put(item)
}
