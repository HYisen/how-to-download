package main

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
)

func destination(count int) string {
	return "./item_" + strconv.Itoa(count)
}

func BenchmarkSuit(b *testing.B) {
	sizes := []string{"128KiB", "2MiB", "8MiB", "32MiB", "512MiB"}
	// FYI, when I wrote this, the default of io.CopyBuffer is 32*1024,
	// which is also the normal path that would actually be.
	bufSizes := []string{"512B", "2KiB", "8KiB", "32KiB", "256KiB", "1MiB"}
	var count int
	for _, size := range sizes {
		b.Run("size="+size, func(b *testing.B) {
			// Ignore err as the same parser will Fatal later in Download.
			bytes, _ := humanize.ParseBytes(size)

			source := "http://localhost:8086/download?size=" + size
			b.Run("buf=normal", func(b *testing.B) {
				b.SetBytes(int64(bytes))
				for b.Loop() {
					if err := Download(context.Background(), source, destination(count)); err != nil {
						b.Fatal(err)
					}
					count++
				}
			})
			count = cleanup(b, count)

			for _, bufSize := range bufSizes {
				bs, _ := humanize.ParseBytes(bufSize)
				count = 0
				b.Run("buf="+bufSize, func(b *testing.B) {
					b.SetBytes(int64(bytes))
					for b.Loop() {
						if err := DownloadWithBuffer(context.Background(), source, destination(count), int(bs)); err != nil {
							b.Fatal(err)
						}
						count++
					}
				})
				count = cleanup(b, count)
			}
		})
	}
}

func cleanup(b *testing.B, count int) (neoCount int) {
	for i := range count {
		if err := os.Remove(destination(i)); err != nil {
			b.Logf("Failed to cleanup remove created file %s: %v", destination(i), err)
		}
	}
	return 0
}

// noReadFrom is copied from that in package io.
type noReadFrom struct{}

// ReadFrom is copied from that in package io.
func (noReadFrom) ReadFrom(io.Reader) (int64, error) {
	panic("can't happen")
}

// fileWithoutReadFrom is copied from that in package io.
type fileWithoutReadFrom struct {
	noReadFrom
	*os.File
}

func DownloadWithBuffer(ctx context.Context, sourceURL string, destinationPath string, bufSize int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer CloseLogError(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not OK status %s", resp.Status)
	}

	fp, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	// One could image that if I set buf too slow, such as 1 Byte, could slow down the speed.
	// But as I tested, without fileWithoutReadFrom, the speed is irrelevant to bufSize.
	// Because io.CopyBuffer invoke buf platform dependent. On macOS, zero-copy is not supported,
	// thus fp.ReadFrom would fall back to io.Copy, which shadow the bufSize setting.
	w := fileWithoutReadFrom{
		noReadFrom: noReadFrom{},
		File:       fp,
	}
	if _, err = io.CopyBuffer(w, resp.Body, make([]byte, bufSize)); err != nil {
		return err
	}
	return nil
}
