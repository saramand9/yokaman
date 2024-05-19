<h1 align="center">
  yokaman cli
</h1>
<h4 align="center">
  A package for yoka robot types and extensions based on magic numbers
</h4>
<h6 align="center">
  Goroutine safe, extensible, no C bindings
</h6>


<p align="center">
  <a href="https://travis-ci.org/gabriel-vasile/mimetype">
    <img alt="Build Status" src="https://travis-ci.org/gabriel-vasile/mimetype.svg?branch=master">
  </a>
  <a href="https://pkg.go.dev/github.com/gabriel-vasile/mimetype">
    <img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/gabriel-vasile/mimetype.svg">
  </a>
  <a href="https://goreportcard.com/report/github.com/gabriel-vasile/mimetype">
    <img alt="Go report card" src="https://goreportcard.com/badge/github.com/gabriel-vasile/mimetype">
  </a>
  <a href="https://codecov.io/gh/gabriel-vasile/mimetype">
    <img alt="Code coverage" src="https://codecov.io/gh/gabriel-vasile/mimetype/branch/master/graph/badge.svg?token=qcfJF1kkl2"/>
  </a>
  <a href="LICENSE">
    <img alt="License" src="https://img.shields.io/badge/License-MIT-green.svg">
  </a>
</p>

## Features
- 支持对协议数据成功/失败/无响应的统计
- 支持对协议数据中位数，TP90, TP95, TP99, 平均值等实时计算
- 支持协议数据自动备份
- 支持本地备份文件上传解析

## Install
```bash
go get github.com/saramand9/yokaman
```

## Usage
```go
	yokacli := yokaman.YoKaManCli()
	yokacli.SetTestInfo(1) //设置testid, 所有数据是按照testid来区分计算的

	yokacli.SetMetricsSvrAddr("172.25.0.1") //设置服务器ip

	err := yokacli.Start() //启动数据上报客户端，在后台会启动线程上传
	if err != nil {
		return
	}

	for i := 0; i < 10; i++ {
		req := yokaman.ReqMetrics{
			Trans:    "login",
			Reqtime:  time.Now().UnixMilli(),
			Resptime: time.Now().UnixMilli(),
			Code:     yokaman.SUCCESS,
			Robotid:  0,
		}
		yokacli.StatReqMetrics(req)
		time.Sleep(time.Second)
	}
	return
```
See the [runnable Go Playground examples](https://pkg.go.dev/github.com/gabriel-vasile/mimetype#pkg-overview).

## FAQ
**Q:** ReqTime/Resptime 填什么？

**A:** ReqTime/Resptime 分别填的是请求发送和收到相应时刻的时间戳，精确到毫秒



**Q:** testid 是什么？ 怎么填？

**A:** 每次测试对应的唯一id，即testid。 平台统计TP90, QPS的时间段，也以testid作为区分。  一般情况下，每次启动机器人都应该重新设置testid，

但是仅是为了协议调试， 尚且不在乎TP90等统计值，也可以暂时不修改。



## Structure
**mimetype** uses a hierarchical structure to keep the MIME type detection logic.
This reduces the number of calls needed for detecting the file type. The reason
behind this choice is that there are file formats used as containers for other
file formats. For example, Microsoft Office files are just zip archives,
containing specific metadata files. Once a file has been identified as a
zip, there is no need to check if it is a text file, but it is worth checking if
it is an Microsoft Office file.

To prevent loading entire files into memory, when detecting from a
[reader](https://pkg.go.dev/github.com/gabriel-vasile/mimetype#DetectReader)
or from a [file](https://pkg.go.dev/github.com/gabriel-vasile/mimetype#DetectFile)
**mimetype** limits itself to reading only the header of the input.
<div align="center">
  <img alt="structure" src="https://github.com/gabriel-vasile/mimetype/blob/420a05228c6a6efbb6e6f080168a25663414ff36/mimetype.gif?raw=true" width="88%">
</div>

## Performance
Thanks to the hierarchical structure, searching for common formats first,
and limiting itself to file headers, **mimetype** matches the performance of
stdlib `http.DetectContentType` while outperforming the alternative package.

```bash
                            mimetype  http.DetectContentType      filetype
BenchmarkMatchTar-24       250 ns/op         400 ns/op           3778 ns/op
BenchmarkMatchZip-24       524 ns/op         351 ns/op           4884 ns/op
BenchmarkMatchJpeg-24      103 ns/op         228 ns/op            839 ns/op
BenchmarkMatchGif-24       139 ns/op         202 ns/op            751 ns/op
BenchmarkMatchPng-24       165 ns/op         221 ns/op           1176 ns/op
```

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md).
