# nginx-log-analyzer

A small CLI utility for analyzing Nginx logs.  

---

## Features

- Input: local path with glob (`logs/**/2024-08-31*`) or single **URL**
- Optional time range or special value filters: `--from`, `--to` in **ISO8601** and `--filter-field`, `--filter-value`
- Output formats: `--format markdown|adoc`
- Stats in **one pass** (streaming, without loading whole file):
  - total requests
  - top requested resources
  - most frequent HTTP status codes
  - average response size
  - **95th percentile** of response size


## Build & Test (Makefile)

```bash
make build      
make test      
``` 
