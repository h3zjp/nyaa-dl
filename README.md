# nyaa-dl

`nyaa-dl` is a command line program to download torrents files from nyaa.

**Don't use this for illegal download purposes. I'm not responsible for any trouble.**

## Installation

Please download commands from [Releases Page](https://github.com/JohnRabin2357/nyaa-dl/releases).

- Windows (64bit): `nyaa-dl_x.x.x_Windows_x86_64.exe`
- Windows (32bit): `nyaa-dl_x.x.x_Windows_i386.exe`

If you want to build it yourself, run `make` or `go build -o ./nyaa-dl.out cmd/nyaa-dl/main.go`.

## Usage

See `./nyaa-dl.exe --help` for more usage.

### Configuration

First, prepare a json configuration file, e.g. `config.json`.

```config.json
{
    "targets": [
        {
            "title": "Art - Doujinshi",
            "domain": "nyaa.example.com",
            "requiredDownloads": 6000,
            "maxPage": 2,
            "category": "1_2",
            "query": "MyFavoriteArtist",
            "trustedOnly": true
        }
    ]
}
```

You can specify multiple targets, and a target speficitaions are here.

- title: Free title.
- domain: The domain of nyaa. e.g. `nyaa.example.com` from `https://nyaa.example.com`.
- requiredDownloads: Minimum number of downloads threshold to download.
- maxPage: Max page to check torrent files.
- category (Optional): The value of `c` from `https://nyaa.example.com/?c=1_2`.
- query (Optional): The value of `q` from `https://nyaa.example.com/?q=YourSearchQuery`.
- trustedOnly: The flag whether is downloading from only trusted user uploaded files.

### Running

For example, you can exec nyaa-dl like `./nyaa-dl.exe --config="./config.json" --output="./my-torrents"`.\
This command crawls nyaa by `./config.json` and saves torrent files to `./my-torrents`. When executed it a second time, skips to download already downloaded torrent files. So basically you don't have to remove torrent files in `./my-torrents`.
