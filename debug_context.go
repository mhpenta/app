package app

import (
	"context"
	"fmt"
	"sync"
)

type DebugContext struct {
	context.Context
	mu   sync.Mutex
	data map[interface{}]interface{}
}

func (d *DebugContext) WithValue(key, val interface{}) *DebugContext {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.data == nil {
		d.data = make(map[interface{}]interface{})
	}
	d.data[key] = val

	return &DebugContext{
		Context: context.WithValue(d.Context, key, val),
		data:    d.data,
	}
}

func (d *DebugContext) PrintValues() {
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Println("Context values - DebugContext")
	for k, v := range d.data {
		fmt.Println("Key:", k, "Value:", v)
	}
}
