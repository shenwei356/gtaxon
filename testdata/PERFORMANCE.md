## Performance

### gi2taxid

- Platform: Fedora Linux (4.3.5-300.fc23.x86_64), RAM 12G, SSD

- Dataset:

    1. gi_small (~250)

            zcat gi_taxid_prot.dmp.gz | cut -f 1 | awk '{if(FNR%1000001==0){print $1}}' | shuf | gzip -c  > gi_small.gz

    2. gi_medium (~25K)

            zcat gi_taxid_prot.dmp.gz | cut -f 1 | awk '{if(FNR%10002==0){print $1}}' | shuf | gzip -c  > gi_medium.gz

    3. gi_large (~2.5M)

            zcat gi_taxid_prot.dmp.gz | cut -f 1 | awk '{if(FNR%103==0){print $1}}' | shuf | gzip -c > gi_large.gz

- Command:

    1. local query

            gtaxon cli local -t gi_taxid_prot -f gi_small.gz

    2. remote query (actually, it's local host here)

            gtaxon server &
            gtaxon cli remote -t gi_taxid_prot -f gi_small.gz

- Result:

    Note: 1) bolt database utilizes the operating system's page cache,
    so repeat queries are faster than the first query. 2) "remote query" actually is from local host 
    with minimum network latency

| dataset        | local query     | remote query     | remote query (repeated) |
|----------------|-----------------|------------------|-------------------------|
| small (0.25K)  |  0.013 s        |   0.013 s        |  0.009s                 |
| medium (25K)   |  0.38 s         |   0.57 s         |  0.178s                 |
| large (2.5M)   |  17 s           |   1min 38s       |  20 s                   |
