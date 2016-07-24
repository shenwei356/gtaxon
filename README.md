# gTaxon

## Introduction

gTaxon - a fast cross-platform NCBI taxonomy data querying tool,
with cmd client and REST API server for both local and remote server.
[http:///github.com/shenwei356/gtaxon](http:///github.com/shenwei356/gtaxon)

## Supporting querying types

|    Query type    |   Function                               | Local/Remote |
|------------------|------------------------------------------|--------------|
|   gi_taxid_nucl  |   query TaxId by Gi (nucl)               |  Both        |
|   gi_taxid_prot  |   query TaxId by Gi (prot)               |  Both        |
|   taxid2taxon    |   query Taxon by TaxId                   |  Remote      |
|   name2taxid     |   query TaxId by Name                    |  Remote      |
|   lca            |   query Lowest Common Ancestor by TaxIds |  Remote      |

## Features

- Easy to install. **Only ONE single executable binary file**.
 No scared source compilation, installing extra packages,
 configuring environment variables
- **Cross platform**. gTaxon is implemented in [golang](https://golang.org).
Executable binary files for most popular operating system (Linux, Mac OS X,
Windows, \*BSD ) are available.
See [Release](https://github.com/shenwei356/gtaxon/releases) page.
- Supporting querying from **both LOCAL and REMOTE server** by REST API,
 which is also easily called by various clients of other languages.
 gTaxon has command-line client `gtaxon cli local` for local query and
 `gtaxon cli remote` for remote query.
- **Fast**. See Section Performance.

## Performance

### gi2taxid

[Detail](https://github.com/shenwei356/gtaxon/blob/master/testdata/PERFORMANCE.md)

Note: 1) bolt database utilizes the operating system's page cache,
so repeat queries are faster than the first query. 2) "remote query" actually is from local host
with minimum network latency

| dataset        | local query     | remote query     | remote query (repeated) |
|----------------|-----------------|------------------|-------------------------|
| small (0.25K)  |  0.013 s        |   0.013 s        |  0.009s                 |
| medium (25K)   |  0.38 s         |   0.57 s         |  0.178s                 |
| large (2.5M)   |  17 s           |   1min 38s       |  20 s                   |

## Download && Install

Steps:

1. Just download and uncompress the executable binary files of your operating system from  [Release](https://github.com/shenwei356/gtaxon/releases) page.

2. Rename it to `gtaxon.exe` (for Windows) or `gtaxon` (for other operating systems) for convenience, and then run it in command-line interface, no compilation, no dependencies.

You can also add the directory of the executable file to environment variable `PATH`, so you can run `gtaxon` anywhere.

1. For windows, the simplest way is copy it to ` C:\WINDOWS\system32`.

2. For Linux, simply copy it to `/usr/local/bin` or add the path of gtaxon to environment variable `PATH`:

        chmod a+x /PATH/OF/GTAXON/gtaxon
        echo export PATH=\$PATH:/PATH/OF/GTAXON >> ~/.bashrc

## Usage

### Loading data to database

1. Initializing database.

        gtaxon db init

2. Importing data

    Supported file types includes:

        ================================================
          data type                  files
        ------------------------------------------------
        gi_taxid_nucl          gi_taxid_nucl.dmp.gz
        gi_taxid_prot          gi_taxid_prot.dmp.gz
        nodes                  nodes.dmp
        names                  names.dmp
        divisions              division.dmp
        gencodes               gencode.dmp
        ================================================

    For gi2taxid

        # ~ 16 min for me
        gtaxon db import -f -t gi_taxid_prot gi_taxid_prot.dmp.gz

    For taxon query

        gtaxon db import -f -t nodes nodes.dmp
        gtaxon db import -f -t names names.dmp
        gtaxon db import -f -t divisions division.dmp
        gtaxon db import -f -t gencodes gencode.dmp

### Querying from local (Only for gi2taxid)

- few queries

        gtaxon cli local -t gi_taxid_prot 139299181 139299182

- from file

        gtaxon cli local -t gi_taxid_prot -f gi_list_file

### Querying from remote server

1. Starting server

        gtaxon server

2. Query TaxId by Gi (gi_taxid_nucl or gi_taxid_prot)

    - few queries

            gtaxon cli remote -t gi_taxid_prot 139299181 139299182

    - from files

            gtaxon cli remote -H 192.168.1.101 -P 8080 -t gi_taxid_prot -f gi_list_file

3. Query TaxId by Name (name2taxid)

    Limiting name class, using regular expression

        gtaxon cli remote -t name2taxid --use-regexp --name-class "scientific name" sapiens
        [INFO] Query TaxId by Name from host: 127.0.0.1:8080
        sapiens 9606(Homo sapiens),1035824(Trichuris sp. ex Homo sapiens JP-2011),1573476(Homo sapiens/Rattus norvegicus xenograft),324570(Phrynium sapiense),63221(Homo sapiens neanderthalensis),1383439(Homo sapiens/Mus musculus xenograft),741158(Homo sapiens ssp. Denisova),399796(Macrobiotus sapiens),349050(Ficus casapiensis),1131344(Homo sapiens x Mus musculus hybrid cell line),270523(Tetragonula sapiens)

        gtaxon cli remote -t name2taxid --use-regexp --name-class "genbank common name" human mouse
        [INFO] Query TaxId by Name from host: 127.0.0.1:8080
        human   121226(Pediculus humanus capitis),121225(Pediculus humanus),51028(Enterobius vermicularis),121224(Pediculus humanus corporis),433352(Diplogonoporus grandis),36087(Trichuris trichiura),115427(Dermatobia hominis),9606(Homo sapiens)
        mouse   42410(Peromyscus eremicus),1595964(Apomys sacobianus),10105(Mus minutoides),221913(Pseudomys hermannsburgensis),240587(Thalpomys cerradensis),409025(Peromyscus melanocarpus) ...

4. Query Taxon by TaxId (taxid2taxon)

        gtaxon cli remote -t taxid2taxon 9
        # result is similar with result of example 5)


5. Query Lowest Common Ancestor by TaxIds (lca)

        gtaxon cli remote -t lca 9606,63221
        [INFO] Query LCA by TaxIds from host: 127.0.0.1:8080
        Query TaxIDs: 9606,63221
        Taxon: {
          "TaxId": 9606,
          "ScientificName": "Homo sapiens",
          "OtherNames": [
            {
              "ClassCDE": "authority",
              "DispName": "Homo sapiens Linnaeus, 1758"
            },
            {
              "ClassCDE": "genbank common name",
              "DispName": "human"
            },
            {
              "ClassCDE": "common name",
              "DispName": "man"
            }
          ],
          "ParentTaxId": 9605,
          "Rank": "species",
          "Division": "Primates",
          "GeneticCode": {
            "GCId": 1,
            "GCName": "Standard"
          },
          "MitoGeneticCode": {
            "MGCId": 2,
            "MGCName": "Vertebrate Mitochondrial"
          },
          "Lineage": "cellular organisms; Eukaryota; Opisthokonta; Metazoa; Eumetazoa; Bilateria; Deuterostomia; Chordata; Craniata; Vertebrata; Gnathostomata; Teleostomi; Euteleostomi; Sarcopterygii; Dipnotetrapodomorpha; Tetrapoda; Amniota; Mammalia; Theria; Eutheria; Boreoeutheria; Euarchontoglires; Primates; Haplorrhini; Simiiformes; Catarrhini; Hominoidea; Hominidae; Homininae; Homo",
          "LineageEx": [
            {
              "TaxId": 131567,
              "ScientificName": "cellular organisms",
              "Rank": "no rank"
            },
            {
              "TaxId": 2759,
              "ScientificName": "Eukaryota",
              "Rank": "superkingdom"
            },
            {
              "TaxId": 33154,
              "ScientificName": "Opisthokonta",
              "Rank": "no rank"
            },
            {
              "TaxId": 33208,
              "ScientificName": "Metazoa",
              "Rank": "kingdom"
            },
            {
              "TaxId": 6072,
              "ScientificName": "Eumetazoa",
              "Rank": "no rank"
            },
            {
              "TaxId": 33213,
              "ScientificName": "Bilateria",
              "Rank": "no rank"
            },
            {
              "TaxId": 33511,
              "ScientificName": "Deuterostomia",
              "Rank": "no rank"
            },
            {
              "TaxId": 7711,
              "ScientificName": "Chordata",
              "Rank": "phylum"
            },
            {
              "TaxId": 89593,
              "ScientificName": "Craniata",
              "Rank": "subphylum"
            },
            {
              "TaxId": 7742,
              "ScientificName": "Vertebrata",
              "Rank": "no rank"
            },
            {
              "TaxId": 7776,
              "ScientificName": "Gnathostomata",
              "Rank": "no rank"
            },
            {
              "TaxId": 117570,
              "ScientificName": "Teleostomi",
              "Rank": "no rank"
            },
            {
              "TaxId": 117571,
              "ScientificName": "Euteleostomi",
              "Rank": "no rank"
            },
            {
              "TaxId": 8287,
              "ScientificName": "Sarcopterygii",
              "Rank": "no rank"
            },
            {
              "TaxId": 1338369,
              "ScientificName": "Dipnotetrapodomorpha",
              "Rank": "no rank"
            },
            {
              "TaxId": 32523,
              "ScientificName": "Tetrapoda",
              "Rank": "no rank"
            },
            {
              "TaxId": 32524,
              "ScientificName": "Amniota",
              "Rank": "no rank"
            },
            {
              "TaxId": 40674,
              "ScientificName": "Mammalia",
              "Rank": "class"
            },
            {
              "TaxId": 32525,
              "ScientificName": "Theria",
              "Rank": "no rank"
            },
            {
              "TaxId": 9347,
              "ScientificName": "Eutheria",
              "Rank": "no rank"
            },
            {
              "TaxId": 1437010,
              "ScientificName": "Boreoeutheria",
              "Rank": "no rank"
            },
            {
              "TaxId": 314146,
              "ScientificName": "Euarchontoglires",
              "Rank": "superorder"
            },
            {
              "TaxId": 9443,
              "ScientificName": "Primates",
              "Rank": "order"
            },
            {
              "TaxId": 376913,
              "ScientificName": "Haplorrhini",
              "Rank": "suborder"
            },
            {
              "TaxId": 314293,
              "ScientificName": "Simiiformes",
              "Rank": "infraorder"
            },
            {
              "TaxId": 9526,
              "ScientificName": "Catarrhini",
              "Rank": "parvorder"
            },
            {
              "TaxId": 314295,
              "ScientificName": "Hominoidea",
              "Rank": "superfamily"
            },
            {
              "TaxId": 9604,
              "ScientificName": "Hominidae",
              "Rank": "family"
            },
            {
              "TaxId": 207598,
              "ScientificName": "Homininae",
              "Rank": "subfamily"
            },
            {
              "TaxId": 9605,
              "ScientificName": "Homo",
              "Rank": "genus"
            }
          ]
        }



## Configuration file for Convenience

Default config file is: `$HOME/.gtaxon.yaml`

This is useful when querying from remote server,
we could type few words by saving flags like host and port to config file.

See https://github.com/ogier/pflag

## REST APIs

1. gi2taxid

        http://127.0.0.1:8080/gi2taxid?db=gi_taxid_prot&gi=139299191111&gi=139299181&gi=139299175

2. name2taxid

        http://localhost:8080/name2taxid?regexp=true&class=genbank+common+name&name=human&name=mouse

3. taxid2taxon

        http://localhost:8080/taxid2taxon?taxid=9906&taxid=2

4. lca

        http://localhost:8080/lca?taxids=9606,63221&taxids=1,2


You can also write client in your favorite programming language.

## Implement details

API reference: [godoc](https://godoc.org/github.com/shenwei356/gtaxon/taxon)

- Programming language: [Go](https://golang.org)
- Database: [bolt](https://github.com/boltdb/bolt), an embedded key/value database for Go
- Web server: [gin](https://github.com/gin-gonic/gin), a fast HTTP web framework written in Go

## Caveats

- 64bit operating system is better.
- `bolt` database utilizes the operating system's page cache, larger virtual memory is better.
- Database file size is 16G after loading gi_taxid_prot.dmp.gz
- About 1.5G RAM usage after starting server
