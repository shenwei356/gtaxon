# gTaxon

## Introduction

gTaxon - a fast cross-platform NCBI taxonomy data querying tool,
with cmd client ans REST API server for both local and remote server.
http:///github.com/shenwei356/gtaxon

## Supporting querying types

- gi2taxid (gi_taxid_nucl and gi_taxid_prot)

## Features

- Easy to install. **Only ONE single executable binary file**.
 No scared source compilation, installing extra packages,
 configuring environment variables
- **Cross platform**. gTaxon is implemented in [golang](https://golang.org). Executable binary files for most popular operating system (Linux, OS X, Windows, \*BSD ) are available. See [Release](https://github.com/shenwei356/gtaxon/releases) page.
- Supporting querying from **both LOCAL and REMOTE server** by REST API, which is also easily called by various clients.
- **Fast**. See Section Performance.

## Performance

Note: 1) bolt database utilizes the operating system's page cache,
so repeat queries are faster than the first query. 2) "remote query" actually is from local host.

| dataset        | local query (s) | remote query (s) |
|----------------|-----------------|------------------|
| small (0.25K)  |  0.015          |   0.007          |
| medium (25K)   |  0.018          |   0.007          |
| large (2.5M)   |  0.016          |   0.012          |



## Download && Install

1. Just download the executable binary files of your operating system from  [Release](https://github.com/shenwei356/gtaxon/releases) page.

2. Rename it to `gtaxon.exe` (for Windows) or `gtaxon` (for other operating systems) for convenience, and then run it in command-line interface, no compilation, no dependencies.

You can also add the directory of the executable file to environment variable `PATH`, so you can run gtaxon anywhere.

1. For windows, the simplest way is copy it to ` C:\WINDOWS\system32`.

2. For Linux, simply copy it to `/usr/local/bin` or add the path of gtaxon to environment variable `PATH`:

        chmod a+x /PATH/OF/GTAXON/gtaxon
        echo export PATH=\$PATH:/PATH/OF/GTAXON >> ~/.bashrc

## Usage

### Loading data to database

1. Initializing database.

        gtaxon db init

2. Importing data

        # ~ 20 min for me
        gtaxon db import -f -t gi_taxid_prot gi_taxid_prot.dmp.gz

### Querying from local

1. few queries

        gtaxon cli local -t gi_taxid_prot 139299181 139299182

2. from file

        gtaxon cli local -t gi_taxid_prot -f gi_list_file

### Querying from remote server

1. Starting server

        gtaxon server

2. Querying

        gtaxon cli remote -H 192.168.1.101 -P 8080 -t gi_taxid_prot -f gi_list_file

## Errors checking

// TODO

## Configuration for Convenience

// TODO

Default config file is: `$HOME/.gtaxon.yaml`

This is useful when querying from remote server,
we could type few words by writing flags like host and port to config file

## REST APIs

Example:

`http://127.0.0.1:8080/gi2taxid?db=gi_taxid_prot&gi=139299191111&gi=139299181&gi=139299175`

## Implement details

see doc: [godoc](https://godoc.org/github.com/shenwei356/gtaxon)

- Programming language: [golang/Go](https://golang.org)
- Database: [bolt](https://github.com/boltdb/bolt), an embedded key/value database for Go
- Web server: [gin](https://github.com/gin-gonic/gin), a fast HTTP web framework written in Go

## Caveats

- database file size is 16G after loading gi_taxid_prot.dmp.gz
- 64bit operating system is better