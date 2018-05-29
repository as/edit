package edit

import (
	"bytes"
	"testing"

	"github.com/as/text"
)

var KB128 = bytes.Repeat([]byte("a"), 1024*128)

func BenchmarkChange128KBto64KB(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	cmd, err := Compile(",x,a,c,aa,")
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkChange128KBto128KB(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,a,c,b,")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkChange128KBto128KBNest4x2x1(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,aaaa,x,aa,x,a,c,b,")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkChange128KBto128KBx16x4x1(b *testing.B) {
	var bufs []text.Editor
	for i := 0; i < b.N; i++ {
		buf, _ := text.Open(text.BufferFrom(append([]byte{}, KB128...)))
		bufs = append(bufs, buf)
	}
	cmd, err := Compile(",x,aaaaaaaaaaaaaaaa,x,aaaa,x,a,c,b,")
	if err != nil {
		b.Fatalf("failed: %s\n", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.Run(bufs[i])
	}
	b.StopTimer()
}

func BenchmarkDelete128KB(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",d")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkDelete128KBx64(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,.{64},d")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkDelete128KBx8(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,.......,d")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkDelete128KBx1(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,.,d")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}
func BenchmarkDelete256KBx1(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,.,d")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}
func BenchmarkDelete512KBx1(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Delete(0, buf.Len())
		buf.Insert(KB128, 0)
		buf.Insert(KB128, 0)
		buf.Insert(KB128, 0)
		buf.Insert(KB128, 0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,.,d")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}
