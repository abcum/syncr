# Syncr

Syncr is a library for storage of append-only log data on local or remote storage.

[![](https://img.shields.io/circleci/token/_____/project/abcum/syncr/master.svg?style=flat-square)](https://circleci.com/gh/abcum/syncr) [![](https://img.shields.io/badge/status-alpha-ff00bb.svg?style=flat-square)](https://github.com/abcum/syncr) [![](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/abcum/syncr) [![](https://goreportcard.com/badge/github.com/abcum/syncr?style=flat-square)](https://goreportcard.com/report/github.com/abcum/syncr) [![](https://img.shields.io/badge/license-Apache_License_2.0-00bfff.svg?style=flat-square)](https://github.com/abcum/syncr) 

#### Features

- Append-only data storage
- Reading and writing of data
- Local and remote streaming storage
- Transparent rotation of append-only files
- Thread safe, for use by multiple goroutines
- Append-only writing to storage using io.Writer
- In-order reading of entire storage using io.Reader
- Ability to buffer writes, or sync writes immediately	
- Write to and read from a directory of log files as if it was one big file
- Support for append-only files locally, and in S3, GCS, RiakCS, CephFS, SeaweedFS

#### Installation

```bash
go get github.com/abcum/syncr
```
