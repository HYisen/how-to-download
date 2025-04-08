package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var src = flag.String("src", "http://localhost:8086/download?size=500MiB", "Source URL")
var dst = flag.String("dst", "./item", "Destination FilePath")

func main() {
	flag.Parse()
	start := time.Now()
	if err := Download(context.Background(), *src, *dst); err != nil {
		log.Fatal(err)
	}
	log.Printf("Download took %v", time.Since(start))
}

func CloseLogError(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Warn("close fail", "err", err)
	}
}

func Download(ctx context.Context, sourceURL string, destinationPath string) error {
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
	if _, err = fp.ReadFrom(resp.Body); err != nil {
		return err
	}
	return nil
}
