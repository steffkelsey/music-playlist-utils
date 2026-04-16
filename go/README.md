# Music Utils Go Tools


## Example Workflow

I find it best to NOT work with your files in place when you want to 
organize your music files while keeping playlists up to date.

Starting from a library where everything is unencrypted and tagged properly...

Export your playlist and linked music files into a separate directory.

First, dry-run:  
```bash
./music-utils playlist export --input-file $HOME/Music/tmp/A\ PT\ playlist.m3u -o $HOME/Music/to-repair -d | jq
```

Then, run it for real:  
```bash
./music-utils playlist export --input-file $HOME/Music/tmp/A\ PT\ playlist.m3u -o $HOME/Music/to-repair
```

Next, organize the music file by tag while exporting a report of the 
the source and destination of each file moved.

First, dry-run:
```bash
./music-utils organize -i $HOME/Music/to-repair -o $HOME/Music/to-repair -d | jq
```

Then, the real thing:
```bash
./music-utils organize -i $HOME/Music/to-repair -o $HOME/Music/to-repair
```

Next, we have broken the playlist by moving the music files. You can verify this by 
validating the playlist:  
```bash
./music-utils playlist validate -i $HOME/Music/to-repair | jq
```

Sample output:   
```json
{
  "valid": [],
  "invalid": [
    {
      "path": "/home/steff/Music/to-repair/A PT playlist.m3u",
      "reason": "One or more bad paths",
      "badPaths": [
        "/home/steff/Music/to-repair/01 TONY.m4a",
        "/home/steff/Music/to-repair/01 Sandstorm.m4a",
        "/home/steff/Music/to-repair/02 Born This Way.m4a",
        "/home/steff/Music/to-repair/03 Opposite of Adults.m4a",
        "/home/steff/Music/to-repair/1-04 Firework.m4a",
        "/home/steff/Music/to-repair/Somebody To Love Remix.mp3",
        "/home/steff/Music/to-repair/05 Shutterbugg (Ft. Cutty).mp3",
        "/home/steff/Music/to-repair/01 Holding On (When Love Is Gone).m4a",
        "/home/steff/Music/to-repair/10 I'm A Sucker for Your Love.m4a",
        "/home/steff/Music/to-repair/05 Acceptable In the 80's.m4a"
      ]
    }
  ]
}
```

Finally, we need to repait the playlist using the report exprted during the
organize step.

First, a dry-run:  
```bash
./music-utils playlist repair -i $HOME/Music/to-repair -c $HOME/Music/to-repair -d | jq
```

Then, the real thing:   
```bash
./music-utils playlist repair -i $HOME/Music/to-repair -c $HOME/Music/to-repair | jq
```

Output:  
```json
{
  "config": [
    "/home/steff/Music/to-repair/organized.json"
  ],
  "skipped": [],
  "repaired": [
    "/home/steff/Music/to-repair/A PT playlist.m3u"
  ],
  "failed": []
}
```
