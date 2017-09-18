package edit

import (
	"bytes"
	"testing"

	"github.com/as/text"
)


func BenchmarkChange1KBto2KB(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++{
		buf.Delete(0, buf.Len())
		buf.Insert(bytes.Repeat([]byte("a"), 1024),0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,a,c,aa,")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkChange1KBto1KB(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++{
		buf.Delete(0, buf.Len())
		buf.Insert(bytes.Repeat([]byte("a"), 1024),0)
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

func BenchmarkChange1KBto1KBNest4x2x1(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++{
		buf.Delete(0, buf.Len())
		buf.Insert(bytes.Repeat([]byte("a"), 1024),0)
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

func BenchmarkChange1KBto1KBx16x4x1(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++{
		buf.Delete(0, buf.Len())
		buf.Insert(bytes.Repeat([]byte("a"), 1024),0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,aaaaaaaaaaaaaaaa,x,aaaa,x,a,c,b,")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkChange1KBto512B(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++{
		buf.Delete(0, buf.Len())
		buf.Insert(bytes.Repeat([]byte("aa"), 512),0)
		buf.Select(0, 0)
		cmd, err := Compile(",x,aa,c,a,")
		if err != nil {
			b.Fatalf("failed: %s\n", err)
		}
		b.StartTimer()
		cmd.Run(buf)
		b.StopTimer()
	}
}

func BenchmarkDelete1KB(b *testing.B) {
	buf, _ := text.Open(text.NewBuffer())
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++{
		buf.Delete(0, buf.Len())
		buf.Insert(bytes.Repeat([]byte("a"), 1024),0)
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
