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
