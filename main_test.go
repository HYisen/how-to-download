package main

import (
	"context"
	"github.com/dustin/go-humanize"
	"os"
	"testing"
)

func BenchmarkNormal(b *testing.B) {
	sizes := []string{"128KiB", "2MiB", "8MiB", "32MiB", "512MiB"}
	destination := "./item"
	for _, size := range sizes {
		b.Run(size, func(b *testing.B) {
			// Ignore err as the same parser will Fatal later in Download.
			bytes, _ := humanize.ParseBytes(size)
			b.SetBytes(int64(bytes))

			source := "http://localhost:8086/download?size=" + size

			for b.Loop() {
				if err := Download(context.Background(), source, destination); err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	if err := os.Remove(destination); err != nil {
		b.Logf("Failed to cleanup remove created file %s: %v", destination, err)
	}
}
