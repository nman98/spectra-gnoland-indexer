# Requirements

The indexer can run on one machine while the database is located on another machine. You can also use
Tiger Data (the TimescaleDB cloud edition), a managed service by Tiger Data, the company behind TimescaleDB.

However, for ease of understanding, throughout this documentation we assume that the indexer and the database are located on the same machine.

## Hardware requirements

To run the indexer you need to have the following system requirements:

Minimum system requirements:

- 2vCPUs
- 8GB RAM

Recommended system requirements:

- 4vCPUs
- 16GB RAM

The indexer could probably run on ARM64 architecture but it is not tested yet. So stick with the x86_64
architecture. There shouldn't be any major difference but if you really want to run it on ARM64 you might need to compile the indexer from the source code.

It can work on HDD for development but for any bigger projects where write and read
speeds are important you might needs SSD depending on what you are doing. The SATA SSD is a good choice for the
most part but you can also use the NVMe SSD for better performance.

For the storage capacity it depends. Given that the indexer is still in development it may vary and it can't be said for certain. For Gnoland Testnet 13 at arround 500K blocks and 250K transactions it used a 
bit over 1.2 GB. This should give you some sort of image of how much space you might need.

For the RAM and CPU it kinda depends but for now this is a good starting point. As the database size 
grows, you might need to increase the RAM and CPU.

Additional info for RAM. It is not that simple for more info look at [tiger data documentation](https://docs.tigerdata.com/use-timescale/latest/hypertables/improve-query-performance/).

To make queries and compression efficient you need to have the necessary amount of RAM. This is kinda hard to calculate but according to the Tiger Data documentation:

```text
The default chunk interval is 7 days. However, best practice is to set chunk_interval so that prior to processing, 
the indexes for chunks currently being ingested into fit within 25% of main memory. For example, on a system with 
64 GB of memory, if index growth is approximately 2 GB per day, a 1-week chunk interval is appropriate. If index 
growth is around 10 GB per day, use a 1-day interval.
```

So this is a bit hard to calculate. For example the Spectra cosmos indexer works in a similar way. At the testing
phase on Osmosis the 100K of block data took about 400MB-700MB of storage space. Now not all of the data is indexed
but for the sake of example let's say it is 700MB. The Osmosis has around 1s block production rate so this is 2
days and maybe around 2/3 of one day. Let's assume that for the week it will take arround 2GB. If Gnoland ever
reached these amount of data you might need to increase the RAM size or modify the chunk interval. But this is if
the indexed data was 2GB. Realistically this is not the case and the indexer data would need a lot of indexed
data.

Good rule of thumb is if the queries are slow increase the RAM size or modify the chunk interval or look at the
queries and see if they are efficient.

8 GB of RAM is a good starting point. 16 GB should be enough for most cases. Depending on how popular the chain is
and how much data is indexed you might need to increase the RAM size.

## Software and OS requirements

The following software and OS requirements are required to run the indexer:

- go 1.26.1
- TimescaleDB 2.18 or higher but with PostgresSQL 16 or higher
- OS: Linux, anything based on Debian(Ubuntu, Mint, etc.) or RHEL(CentOS stream, Rocky Linux, etc.) should work, openSUSE also ok
- Docker (optional)
