package core

import (
	"testing"
)

func BenchmarkMessageBus_Action(b *testing.B) {
	c, _ := New()
	c.RegisterAction(func(c *Core, msg Message) error {
		return nil
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.ACTION("test")
	}
}

func BenchmarkMessageBus_Query(b *testing.B) {
	c, _ := New()
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "result", true, nil
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = c.QUERY("test")
	}
}

func BenchmarkMessageBus_Perform(b *testing.B) {
	c, _ := New()
	c.RegisterTask(func(c *Core, t Task) (any, bool, error) {
		return "result", true, nil
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = c.PERFORM("test")
	}
}
