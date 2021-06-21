# nyaa-dl

`nyaa-dl` is a command line program to download torrents files from nyaa.

**Don't use this for illegal download purposes.**
**I'm not responsible for any trouble.**

## Installation

Download the command from [Releases Page](https://github.com/JohnRabin2357/nyaa-dl/releases).


## Usage

### Configuration

First, prepare a json configuration file.

```config.json
{
    "targets": [
        {
            "title": "Art - Doujinshi",
            "requiredDownloads": 6000,
            "maxPage": 2,
            "category": "1_2",
            "query": ""
        }
    ]
}
```

You can specify multiple targets, and a target speficitaions are here.

- title: free title.
- requiredDownloads: Minimum number of downloads threshold to download.
- maxPage: Max page to crawl.
- category (Optional): `c` of `https://nyaa.example.com/?c=1_2`.
- query (Optional): `q` of `https://nyaa.example.com/?q=YourSearchQuery`.

### How to run

Example:`./nyaa-dl.exe --config="./config.json" --domain="nyaa.example.com" --output="./my-torrents"`.

See `./nyaa-dl.exe --help` for more usage.
