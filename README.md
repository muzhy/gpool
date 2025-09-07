# gpool

`gpool` is a simple, generic, and type-safe object pool for Go, built as a wrapper around the standard library's `sync.Pool`. It leverages Go 1.18+ generics to provide a more convenient and safer API for pooling and reusing objects.

## Features

*   **Type-Safe**: Eliminates the need for type assertions (`.(T)`) when getting objects from the pool.
*   **Simple API**: A minimal and intuitive API with `New`, `Get`, and `Put` methods.
*   **Thread-Safe**: Inherits the concurrency safety of the underlying `sync.Pool`.
*   **Panic Safety**: Gracefully handles cases where the pool's `New` function might return `nil`, preventing panics by returning the zero value for the type.

## Installation

To use `gpool` in your project, you can simply copy the `pool.go` file into your project or manage it as a local module.

If it were a remote module, you would install it like this:
```bash
go get github.com/muzhy/gpool 
```

## Usage

Using `gpool` is straightforward.

### 1. Create a Pool

First, create a new pool using `gpool.New`. You must provide a `newFunc` that will be called to create a new object when the pool is empty.

```go
import "github.com/your-username/gpool"

// Create a pool for *bytes.Buffer objects.
bufferPool := gpool.New(func() *bytes.Buffer {
    // The New function is called when a new instance is needed.
    return new(bytes.Buffer)
})
```

### 2. Get an Object

Use the `Get()` method to retrieve an object from the pool. If the pool has a reusable object, it will be returned; otherwise, your `newFunc` will be called to create a new one.

```go
buf := bufferPool.Get()
// buf is of type *bytes.Buffer, no type assertion needed.
```

### 3. Put an Object Back

After you are done with the object, return it to the pool using the `Put()` method so it can be reused.

**Important**: It is the user's responsibility to reset the object to a clean state before putting it back in the pool.

```go
buf.WriteString("some temporary data")

// ... do work with buf ...

// Reset the buffer before returning it to the pool.
buf.Reset() 
bufferPool.Put(buf)
```

## Complete Example

Here is a complete example demonstrating the basic usage of `gpool`.

```go
package main

import (
	"bytes"
	"fmt"

	"gpool" // Assuming gpool is in your project
)

func main() {
	// Create a pool of *bytes.Buffer.
	bufferPool := gpool.New(func() *bytes.Buffer {
		fmt.Println("Creating a new buffer.")
		return new(bytes.Buffer)
	})

	// Get a buffer from the pool. This will call the New function.
	buf1 := bufferPool.Get()
	buf1.WriteString("Hello, Pool!")
	fmt.Printf("Buffer 1 content: %s\n", buf1.String())

	// Reset and put the buffer back.
	buf1.Reset()
	bufferPool.Put(buf1)

	// Get a buffer again. This should reuse the one we just put back.
	buf2 := bufferPool.Get()
	fmt.Println("Buffer 2 should be empty.")
	fmt.Printf("Buffer 2 content: '%s'\n", buf2.String())
}
```

