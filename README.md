# How To Download

## Goal

Help me find the proper way to download a file in Go.

## Story

Should I introduce a buffer between `http.Response.Body` and `os.File`? If yes, how large shall it be?

This idea comes from a long ago Java code
[snippet](https://github.com/HYisen/Eta0/blob/master/book/src/main/java/net/alexhyisen/eta/book/Utils.java#L51)
of my project.

As I dug inside the go source code, there is no buffer in downstream, and likely neither in upstream.

In my opinion, it makes sense as context switch between NetworkInterfaceController and app and FileSystem costs.

But I shall benchmark it. And this project is that.

## Usage

### STEP 1 Start Server

```shell
go run ./tools/server
```

If fails like port conflict, fix it yourself.

Then bench in **another** terminal.

### STEP 2 Run Benchmark

A whole suit can take 42s, expect a longer time for run multiple times.

The count 6 is minimal for `benchstat` to conclude confidence, count 10 is so-called minimal in its doc.

```shell
go test -bench=. -count=10 | tee new.txt
```

Now you could stop the server in STEP 1, maybe through Ctrl + C or kill.

### STEP 3 Do Analyse

Install one if you haven't.

```shell
go install golang.org/x/perf/cmd/benchstat@latest
```

Check [the official documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) for RTFM.

See how the filesize and bufSize matters.

```shell
benchstat -col /buf new.txt
```

Or you can have a look over my collected data.

```shell
cd data
benchstat -filter .unit:B/s -table /size -row /buf tmpfs-macOS.txt ssd-macOS.txt tmpfs-linux.txt ssd-linux.txt
```

The memory and SSD of macOS is the cheapest one of that CPU model, all are embedded by Apple.

As for linux, it's 2x DDR5 6000 32GiB and PM871 512GiB as ext4 over LVM mounted on slash.

Notably, macOS is used as dev machine, thus not quite idle while benchmarking, although I have halted the IDE scan.

## Results

### Server

The test method is I input the URL to Firefox address bar and monitor the browser download speed.

| Name      | URL                                   | Speed MiB/s |
|-----------|---------------------------------------|-------------|
| Internet  | link.testfile.org/500MB               | 10          |
| localhost | localhost:8086/download?size=500MiB   | >500        |
| LAN       | 192.168.2.2:8000/download?size=500MiB | 12          |

The SpeedTest reports it's 300 Mbps. Most likely the bottleneck is Wi-Fi.

For 10 GiB file, the speed of localhost is 900 MiB/s, seems the bottleneck is SSD.

### SameFile

Testing on a same destination filepath results in weired result,

as SSD and tmpfs share a same speed, and I can't observe a high disk IO.

Maybe the flush does not work as expected or inode manipulation majors.

Through use varied filename, in the cost of a higher storage consumption,

I could observe a high IO, which allows me to assume I have fixed it.
