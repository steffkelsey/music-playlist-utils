# Music File / Playlist File Utils

The problem I need to solve is that I recently recovered some HDs that have
a ton of music on them from old iTunes libs. The files are not correctly
organized for adding to a Jellyfin server - lots of loose files that are not
in a folder or folders and files that are labeled incorrectly. In addition, the
metadata looks rough and I suspect a bunch of duplicated files! There are some 
m4p files that are stuck in Fairplay encryption (about 775 incl' possible dupes).

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
- [x] Find duplicate files and prompt user to remove them one by one or confirm all with report outputting what was kept and what was deleted
- [x] make sure all files have metadata for Album Title, Track Num, Track Name,
Track Artist
- [x] copy all files that don't have metadata to another root folder preserving subfolders (easier for musicbrainz etc to deal with)
- [x] move files from current location to desired location by metatag data as outlined above
- [x] create a JSON file showing where each file was moved to and from
- [x] for a given m3u and JSON file, update each track with the new location preserving track order
- [x] for each m3u file, show any broken music file paths
- [ ] do a fuzzy search by filename (eg: omit last 3 chars before extension)
- [x] move an m3u file to a new folder and update the tracks so it doesn't break (if using relative paths)

In Python
- [ ] use the exiftool binary to read the metadata for album, artist, track name, etc for each encrypted file
- [ ] Use ytmusicapi to find entire albums of encrypted Music
- [ ] Use ytmusicapi to find individual tracks that match (maybe a Greatest Hits album doesn't exist on YT Music but the tracks exist on other albums)
- [ ] Use yt-dlp to download the tracks

## How to:

1. Get metadata from encrypted music files in json

```bash
exiftool -j <filename>
```

```json
[{
  "SourceFile": "./18 Champion Requiem.m4p",
  "ExifToolVersion": 13.56,
  "FileName": "18 Champion Requiem.m4p",
  "Directory": ".",
  "FileSize": "4.8 MB",
  "FileModifyDate": "2026:04:15 08:57:27-04:00",
  "FileAccessDate": "2026:04:16 14:32:25-04:00",
  "FileInodeChangeDate": "2026:04:15 08:57:27-04:00",
  "FilePermissions": "-rw-r--r--",
  "FileType": "M4P",
  "FileTypeExtension": "m4p",
  "MIMEType": "audio/mp4",
  "MajorBrand": "Apple iTunes AAC-LC (.M4A) Audio",
  "MinorVersion": "0.0.0",
  "CompatibleBrands": ["M4A ","mp42","isom"],
  "MovieHeaderVersion": 0,
  "CreateDate": "2004:10:06 21:24:01",
  "ModifyDate": "2007:01:22 11:23:58",
  "TimeScale": 600,
  "Duration": "0:04:53",
  "PreferredRate": 1,
  "PreferredVolume": "100.00%",
  "PreviewTime": "0 s",
  "PreviewDuration": "0 s",
  "PosterTime": "0 s",
  "SelectionTime": "0 s",
  "SelectionDuration": "0 s",
  "CurrentTime": "0 s",
  "NextTrackID": 2,
  "TrackHeaderVersion": 0,
  "TrackCreateDate": "2004:10:06 21:24:01",
  "TrackModifyDate": "2007:01:22 11:23:58",
  "TrackID": 1,
  "TrackDuration": "0:04:53",
  "TrackLayer": 0,
  "TrackVolume": "100.00%",
  "MatrixStructure": "1 0 0 0 1 0 0 0 1",
  "MediaHeaderVersion": 0,
  "MediaCreateDate": "2004:10:06 21:24:01",
  "MediaModifyDate": "2007:01:22 11:23:58",
  "MediaTimeScale": 44100,
  "MediaDuration": "0:04:53",
  "MediaLanguageCode": "und",
  "HandlerDescription": "soun",
  "Balance": 0,
  "AudioFormat": "drms",
  "AudioChannels": 2,
  "AudioBitsPerSample": 16,
  "AudioSampleRate": 44100,
  "OriginalFormat": "mp4a",
  "SchemeType": "itun",
  "SchemeVersion": 0,
  "SchemeURL": "",
  "UserID": "0x087bb07f",
  "KeyID": "0x00000002",
  "InitializationVector": "24e6f543f3871056ab34f27f65ff0076",
  "ItemVendorID": "0x00000003",
  "Platform": "0x00000000",
  "VersionRestrictions": "0x01010100",
  "TransactionID": "0xc1da9437",
  "MediaFlags": "0x00000001",
  "UserName": "sara sweet rabidoux",
  "HandlerType": "Metadata",
  "HandlerVendorID": "Apple",
  "VolumeNormalization": "10EA 136C 8A74 8C74 1DBFE 1DBFE 7FFF 7FFF 14FD6 8CC9",
  "iTunTool": "0x01068000",
  "Title": "Champion Requiem",
  "Artist": "Mos Def",
  "AlbumArtist": "Mos Def",
  "Composer": "88 Keys",
  "Album": "The New Danger",
  "Genre": "Hip-Hop/Rap",
  "TrackNumber": "18 of 18",
  "DiskNumber": "1 of 1",
  "ContentCreateDate": "2004:10:12 07:00:00Z",
  "PlayGap": "Insert Gap",
  "AppleStoreAccount": "sara_sweetly@yahoo.com",
  "Copyright": "℗ 2004 Geffen Records",
  "AppleStoreCatalogID": 25198278,
  "Rating": "Explicit",
  "ArtistID": 92012,
  "ComposerID": 25198280,
  "AlbumID": 25198055,
  "GenreID": "Music|Hip-Hop/Rap",
  "AppleStoreCountry": "United States",
  "AppleStoreAccountType": "iTunes",
  "MediaType": "Normal (Music)",
  "PurchaseDate": "2007-01-22 16:17:59",
  "CoverArt": "(Binary data 88425 bytes, use -b option to extract)",
  "MediaDataSize": 4693913,
  "MediaDataOffset": 145981,
  "Warning": "Unknown trailer with truncated 'o\\xecn2' data at offset 0x49d9d6",
  "AvgBitrate": "128 kbps"
}]
```

### Test data for comparing Tracks and Albums

```json
{
    "track1": {
        "title":"Master Blaster (Jammin')",
        "artist":"Stevie Wonder",
        "trackNumber": 3,
        "totalTracks": 8,
        "album": "Stevie Wonder's Original Musiquarium I (Reissue)",
        "albumArtist": "Stevie Wonder",
        "durationSeconds": 308
    },
    "track2": {
        "title": "Master Blaster (Jammin')",
        "artist": "Stevie Wonder",
        "trackNumber": 3,
        "totalTracks": 8,
        "album": "Original Musiquarium I",
        "albumArtist": "Stevie Wonder",
        "durationSeconds": 308
    }
}
```
```base64
"eyJ0cmFjazEiOnsidGl0bGUiOiJNYXN0ZXIgQmxhc3RlciAoSmFtbWluJykiLCJhcnRpc3QiOiJTdGV2aWUgV29uZGVyIiwidHJhY2tOdW1iZXIiOjMsInRvdGFsVHJhY2tzIjo4LCJhbGJ1bSI6IlN0ZXZpZSBXb25kZXIncyBPcmlnaW5hbCBNdXNpcXVhcml1bSBJIChSZWlzc3VlKSIsImFsYnVtQXJ0aXN0IjoiU3RldmllIFdvbmRlciIsImR1cmF0aW9uU2Vjb25kcyI6MzA4fSwidHJhY2syIjp7InRpdGxlIjoiTWFzdGVyIEJsYXN0ZXIgKEphbW1pbicpIiwiYXJ0aXN0IjoiU3RldmllIFdvbmRlciIsInRyYWNrTnVtYmVyIjozLCJ0b3RhbFRyYWNrcyI6OCwiYWxidW0iOiJPcmlnaW5hbCBNdXNpcXVhcml1bSBJIiwiYWxidW1BcnRpc3QiOiJTdGV2aWUgV29uZGVyIiwiZHVyYXRpb25TZWNvbmRzIjozMDh9fQ=="
```

```json
{
    "album1": {
        "album": "Totally Different",
		"artist": "some guy",
		"totalDiscs": 6
	},
    "album2": {
        "album": "Album 1",
		"artist": "artist 1",
		"totalDiscs": 1
	}
}
```

```base64
"eyJhbGJ1bTEiOnsiYWxidW0iOiJUb3RhbGx5IERpZmZlcmVudCIsImFydGlzdCI6InNvbWUgZ3V5IiwidG90YWxEaXNjcyI6Nn0sImFsYnVtMiI6eyJhbGJ1bSI6IkFsYnVtIDEiLCJhcnRpc3QiOiJhcnRpc3QgMSIsInRvdGFsRGlzY3MiOjF9fQ=="
```
