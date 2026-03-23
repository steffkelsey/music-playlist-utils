# Music File / Playlist File Utils

The problem I need to solve is that I recently recovered some HDs that have
a ton of music on them from old iTunes libs. The files are not correctly
organized for adding to a Jellyfin server - lots of loose files that are not
in a folder or folders and files that are labeled incorrectly. BUT, all the
metadata looks really good! And, there are some m4p files that are stuck in 
Fairplay encryption (about 775 of them).

## Final state

Folder structure and track naming should be

- Music
     |- Album_Title
               |- Track_Num - Track_Name - Track_Artist.mp3/m4a
     |- World Clique
               |- 02 - Good Beat - Dee-Lite.mp3
               |- 03 - Power of Love - Dee-Lite.mp3
               |- 04 - Try Me On... I'm Very You - Dee-Lite.mp3

And an m3u file containing the same tracks would look like:

```Music/example-playlist.m3u
./World Clique/02 - Good Beat - Dee-Lite.mp3
./World Clique/03 - Power of Love - Dee-Lite.mp3
./World Clique/04 - Try Me On... I'm Very You - Dee-Lite.mp3
```

A JSON file that shows where each file moved to
```json
[
    {
        "old": "src_root/Music/Good Beat.mp3",
        "new": "dst_root/Music/World Clique/02 - Good Beat - Dee-Lite.mp3"
    },
    ...
]
```

## TODO

In Go
- [ ] make sure all files have metadata for Album Title, Track Num, Track Name,
Track Artist
- [ ] move files from current location to desired location as outlined above
- [ ] create a JSON file showing where each file was copied or moved to
- [ ] for a given m3u and JSON file, update each track with the new location preserving track order

In Python
- [ ] Use ytmusicapi to find entire albums of encrypted Music
- [ ] Use ytmusicapi to find individual tracks that match (maybe a Greatest Hits album doesn't exist on YT Music but the tracks exist on other albums)
