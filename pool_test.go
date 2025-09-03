package gpool

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

// TestPool_Basic 测试 Get 和 Put 的基本功能。
func TestPool_Basic(t *testing.T) {
	// 用于追踪 New 函数被调用次数的计数器
	var newCounter int32

	// 为 *bytes.Buffer 创建一个池
	p := New(func() *bytes.Buffer {
		atomic.AddInt32(&newCounter, 1)
		return new(bytes.Buffer)
	})

	// 1. 从空池中获取一个对象
	// New 函数应该被调用一次
	buf1 := p.Get()
	if buf1 == nil {
		t.Fatal("从空池中 Get() 应该返回一个新对象")
	}
	if atomic.LoadInt32(&newCounter) != 1 {
		t.Fatalf("第一次 Get() 时，New() 应该被调用一次, 但实际调用了 %d 次", atomic.LoadInt32(&newCounter))
	}

	// 2. 将对象放回池中
	p.Put(buf1)

	// 3. 再次获取对象
	// 应该复用之前放回的对象，New() 不应再次被调用
	buf2 := p.Get()
	if buf2 == nil {
		t.Fatal("Get() 应该返回一个池化的对象")
	}
	if atomic.LoadInt32(&newCounter) != 1 {
		t.Fatalf("复用对象时，New() 不应该被再次调用, 但实际调用了 %d 次", atomic.LoadInt32(&newCounter))
	}

	// 4. 验证获取的是同一个实例
	if buf1 != buf2 {
		t.Fatal("应该从池中获取到相同的实例")
	}
}

// TestPool_Reset 演示了在将对象放回池之前重置它的重要性。
func TestPool_Reset(t *testing.T) {
	type ResettableObject struct {
		data string
	}

	p := New(func() *ResettableObject {
		return &ResettableObject{}
	})

	// 获取一个对象，修改它，然后在不重置的情况下将其放回。
	obj1 := p.Get()
	obj1.data = "dirty"
	p.Put(obj1)

	// 再次获取对象。它应该是同一个，并且仍然是“脏”的。
	obj2 := p.Get()
	if obj2.data != "dirty" {
		t.Errorf("如果对象没有被重置，它应该保持其状态, 期望 'dirty', 得到 '%s'", obj2.data)
	}

	// 现在，我们来实践良好的卫生习惯。
	// 在放回之前重置对象。
	obj2.data = "" // 手动重置
	p.Put(obj2)

	// 再次获取。它应该是干净的。
	obj3 := p.Get()
	if obj3.data != "" {
		t.Errorf("对象在重置后应该是干净的, 期望 '', 得到 '%s'", obj3.data)
	}
}

// TestPool_Concurrency 测试池在并发使用下的安全性。
func TestPool_Concurrency(t *testing.T) {
	p := New(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})

	// 使用 GOMAXPROCS 个 goroutine 并发地操作池
	numGoroutines := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			buf := p.Get()
			buf.WriteString(fmt.Sprintf("goroutine %d", id))
			buf.Reset() // 在放回前重置
			p.Put(buf)
		}(i)
	}

	wg.Wait()
}

// TestPool_Get_WithNilFromNew 测试当池的 New 函数返回 nil 时 Get 方法的行为。
func TestPool_Get_WithNilFromNew(t *testing.T) {
	t.Run("PointerType", func(t *testing.T) {
		// 为指针类型创建一个池，其 New 函数返回 nil。
		p := New(func() *bytes.Buffer {
			return nil
		})

		// Get 应该返回一个 nil 指针，而不是 panic。
		v := p.Get()
		if v != nil {
			t.Errorf("当 New 返回 nil 时，对于指针类型期望得到 nil，但得到了 %v", v)
		}
	})

	t.Run("ValueType", func(t *testing.T) {
		// 对于值类型，其 New 函数不能返回 nil。
		// 但我们可以通过直接操作内嵌的 sync.Pool 来模拟这种情况，
		// 即将其 New 函数设置为返回 nil。
		// 这可以验证我们的 Get 方法能够防止 `nil.(T)` 的 panic。
		type ValueObject struct {
			X int
		}

		p := New(func() ValueObject {
			// 这个函数实际上不会被调用
			return ValueObject{X: 1}
		})

		// 覆盖内嵌的 sync.Pool 的 New 函数
		p.Pool.New = func() any { return nil }

		// Get 应该返回 ValueObject 的零值，而不是 panic。
		v := p.Get()
		if v.X != 0 {
			t.Errorf("当 New 返回 nil 时，对于值类型期望得到零值，但得到了 %+v", v)
		}
	})
}
