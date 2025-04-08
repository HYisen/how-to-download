package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var addr = flag.String("addr", "localhost:8086", "listen address to serve")
var pool = flag.String("pool", "1MiB", "memory consumption of the pre generated file part")

func main() {
	flag.Parse()

	arena, err := Arena(*pool)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /download", func(writer http.ResponseWriter, request *http.Request) {
		size, err := AskedSize(request)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			if _, err := writer.Write([]byte(err.Error())); err != nil {
				log.Printf("Can not write error response: %v", err)
			}
		}

		log.Println(filename(size) + " Begin")
		writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename(size)))
		writer.Header().Add("Content-Type", "application/octet-stream")
		writer.Header().Add("Content-Length", strconv.FormatUint(size, 10))
		writer.WriteHeader(http.StatusOK)

		quotient, remainder := Divide(size, len(arena))
		for range quotient {
			if _, err := writer.Write(arena); err != nil {
				log.Printf("Can not write payload part: %v", err)
				return
			}
		}
		if _, err := writer.Write(arena[:remainder]); err != nil {
			log.Printf("Can not write payload remain: %v", err)
			return
		}
		log.Println(filename(size) + " End")
		return
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func filename(size uint64) string {
	return strings.ReplaceAll(humanize.IBytes(size), " ", "_") + ".txt"
}

func Arena(poolSize string) ([]byte, error) {
	size, err := humanize.ParseBytes(poolSize)
	if err != nil {
		return nil, err
	}

	content := "DON'T LOOK AT THE STARS.\n"
	mul, add := Divide(size, len(content))
	ret := bytes.Repeat([]byte(content), mul)
	for range add {
		ret = append(ret, '.')
	}
	return ret, nil
}

func Divide(dividend uint64, divisor int) (quotient, remainder int) {
	// avoid silly overflow case
	if divisor <= 0 {
		panic(fmt.Errorf("invalid divisor %d", divisor))
	}
	quotient = int(dividend / uint64(divisor))
	remainder = int(dividend - uint64(divisor)*uint64(quotient))
	return quotient, remainder
}

func AskedSize(request *http.Request) (uint64, error) {
	ret := uint64(1024) // By default, we provide an 1 KiB file.
	if str := request.URL.Query().Get("size"); str != "" {
		num, err := humanize.ParseBytes(str)
		if err != nil {
			return 0, err
		}
		ret = num
	}
	return ret, nil
}
