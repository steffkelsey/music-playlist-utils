# Python CLI for finding and downloading music

## Requirements
Install exiftools  
https://exiftool.org/install.html#Unix

## Usage

To Setup ytmusic api

The oauth flow is broken, so login via Firefox, copy request headers
from `/browse` request.

Create browser.json:
```bash
ytmusic browser
```

Then in the script:  
```Python
from ytmusicapi import YTMusic
ytmusic = YTMusic("browser.json")
```

To search for an album (The New Danger by Mos Def)
```Python
ytmusic.search("The New Danger", "album")
```

Sample output: 
```json
[
  {
    "category": "Albums",
    "resultType": "album",
    "title": "The New Danger",
    "type": "Album",
    "playlistId": "OLAK5uy_mZUV3s95o0wppzvcBghh6qHdaaLZGcE9o",
    "duration": "none",
    "year": "2004",
    "artists": [
      {
        "name": "Mos Def",
        "id": "UChAY_qKqbGyqau8q1gXkzww"
      }
    ],
    "browseId": "MPREb_rI6rCsYe5uM",
    "isExplicit": "true",
    "thumbnails": [
      {
        "url": "https://yt3.googleusercontent.com/-JmHJ3NdvQg8KPs-ofk6F_YEJp6EATZlOFsYBeLQ3DRAnajwXIN4Jb167OQXIhvEllPfuZpSbRPJYsI=w60-h60-l90-rj",
        "width": 60,
        "height": 60
      },
      {
        "url": "https://yt3.googleusercontent.com/-JmHJ3NdvQg8KPs-ofk6F_YEJp6EATZlOFsYBeLQ3DRAnajwXIN4Jb167OQXIhvEllPfuZpSbRPJYsI=w120-h120-l90-rj",
        "width": 120,
        "height": 120
      },
      {
        "url": "https://yt3.googleusercontent.com/-JmHJ3NdvQg8KPs-ofk6F_YEJp6EATZlOFsYBeLQ3DRAnajwXIN4Jb167OQXIhvEllPfuZpSbRPJYsI=w226-h226-l90-rj",
        "width": 226,
        "height": 226
      },
      {
        "url": "https://yt3.googleusercontent.com/-JmHJ3NdvQg8KPs-ofk6F_YEJp6EATZlOFsYBeLQ3DRAnajwXIN4Jb167OQXIhvEllPfuZpSbRPJYsI=w544-h544-l90-rj",
        "width": 544,
        "height": 544
      }
    ]
  },
  {...},
] 
```

We need to match the album with artist. There might be more than one album artist.

Album and artist / album artist will be pulled from exiftool


Plug the `playlistId` into yt-dlp:  
```bash
yt-dlp --cookies ./cookies.txt -P "~/Music/dl" -o "%(album)s/%(autonumber)02d - %(track)s.%(ext)s" -x --audio-format mp3 --add-metadata "https://music.youtube.com/playlist?list=OLAK5uy_mZUV3s95o0wppzvcBghh6qHdaaLZGcE9o"
```
