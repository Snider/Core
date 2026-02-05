package io

import (
	"testing"
)

func BenchmarkMockMedium_Write(b *testing.B) {
	m := NewMockMedium()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Write("test.txt", "some content")
	}
}

func BenchmarkMockMedium_Read(b *testing.B) {
	m := NewMockMedium()
	_ = m.Write("test.txt", "some content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Read("test.txt")
	}
}

func BenchmarkMockMedium_List(b *testing.B) {
	m := NewMockMedium()
	_ = m.EnsureDir("dir")
	for i := 0; i < 100; i++ {
		_ = m.Write("dir/file"+string(rune(i))+".txt", "content")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.List("dir")
	}
}
