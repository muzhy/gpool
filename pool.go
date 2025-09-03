package gpool

import "sync"

// Pool 是一个围绕 sync.Pool 的泛型、类型安全的包装器。
// 它通过嵌入 sync.Pool 来继承其基本行为。
type Pool[T any] struct {
	sync.Pool
}

// New 创建一个新的 Pool。
// 当池为空时，提供的 newFunc 函数将被调用以创建新对象。
//
// 为了获得最佳性能并避免不必要的内存分配，newFunc 最好返回一个指针类型 (*T)。
func New[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		Pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get 从池中获取一个 T 类型的对象，并提供类型安全。
func (p *Pool[T]) Get() T {
	v := p.Pool.Get()
	if v == nil {
		// 如果池返回 nil，安全地返回 T 类型的零值，
		// 避免当 T 是值类型时发生 `nil.(T)` 的 panic。
		var zero T
		return zero
	}
	return v.(T)
}

// Put 将一个 T 类型的对象放回池中。
func (p *Pool[T]) Put(x T) {
	p.Pool.Put(x)
}
