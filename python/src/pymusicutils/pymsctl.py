import base64
import typer
from typing_extensions import Annotated
import os
from .validate import fileOrDirectoryExists
import json
import time
import copy
import subprocess
from ytmusicapi import YTMusic
from bisect import bisect_left

def pymsctl(
    input_file: Annotated[str, typer.Option(help="The json report with albums to search for")],
    output_dir: Annotated[str, typer.Option(help="The directory to size downloaded music files")],
    dry_run: Annotated[bool, typer.Option(help="Dry-run to print json results to stdout")] = False,
    browser_json: Annotated[str, typer.Option(help="Path to the browser.json file from 'ytmusic browser' setup", envvar="BROWSER_JSON")] = "$HOME/browser.json",
    cookies_txt: Annotated[str, typer.Option(help="Path to the cookies.txt from exporting cookies from browser in Netscape format", envvar="COOKIES_TXT")] = "$HOME/cookies.txt"
):
    input_file = os.path.expandvars(input_file)
    output_dir = os.path.expandvars(output_dir)
    browser_json = os.path.expandvars(browser_json)
    cookies_txt = os.path.expandvars(cookies_txt)
    # validate input_file exists
    if fileOrDirectoryExists(input_file) == False:
        print(f"input-file does not exist at {input_file}")
        raise typer.Exit(code=1)
    # validate output_dir exists
    if fileOrDirectoryExists(output_dir) == False:
        print(f"output-dir does not exist at {output_dir}")
        raise typer.Exit(code=1)
    # validate browser.json file exists
    if fileOrDirectoryExists(browser_json) == False:
        print(f"browser.json does not exist at {browser_json}")
        raise typer.Exit(code=1)
    # validate cookies.txt file exists
    if fileOrDirectoryExists(cookies_txt) == False:
        print(f"cookies.txt does not exist at {cookies_txt}")
        raise typer.Exit(code=1)

    # open input file
    with open(input_file, 'r', encoding='utf-8') as f:
        # parse the json into a python object
        exif_report = json.load(f)
    
    # create a tuple for saving ytmusicapi results to download
    playlistIdsToDownload = []
    videoIdsToDownload = []
    # create tuple to report on what was downloaded
    albums = []
    # create a tuple for skipped albums (no good matches etc)
    skipped = []

    # create an authenticated instance of the ytmusicapi
    ytmusic = YTMusic(browser_json)

    # iterate over the albums
    for a in exif_report["albums"]:
        matched = False
        if not dry_run:
            print(f"...matching... '{a['artist']}' - '{a['album']}'")
        #print(f"album: {a["album"]}")
        #print(f"artist: {a["artist"]}")
        #print()
        # on each album, search with ytmusicapi
        album_search_results = ytmusic.search(a["album"], 'albums')
        for r in album_search_results:
            if r['resultType'] != 'album':
                continue
            # we need to query the music-utils go application for each album
            # to see if the match is reasonable enough to continue
            maybeAlbum = get_album_from_response(r)
            score, track_match_indexes = cmp_albums(a, maybeAlbum, True)

            # If the score is above the threshold, get the entire album info from yt
            # 0.5 is pretty good because there is no field for trackCount (totalTracks)
            # at in the returned ablum object in search 
            # making the best possible score a 0.667
            if (score > 0.5):
                # get the complete album info by querying with the browseId
                time.sleep(0.33)
                yt_album = ytmusic.get_album(r['browseId'])
                # update the score and track match indexes
                maybeAlbum = get_album_from_response(yt_album)
                score, track_match_indexes = cmp_albums(a, maybeAlbum, False)

                # Check if there are alternate versions
                if 'other_versions' in yt_album:
                    best_other_score, best_other_track_match_indexes, best_other_album = get_best_other_album(ytmusic, yt_album['other_versions'], a)
                    if best_other_score > score:
                        yt_album = best_other_album
                        score = best_other_score
                        track_match_indexes = best_other_track_match_indexes

                    if score > 0.8:
                        matched = True
                        playlistIdsToDownload.append(yt_album['audioPlaylistId'])
                        print(f"very good match on {a['album']}")
                        removed_matched_tracks(a, track_match_indexes)
                        # store the artist for maybe looking up any missing songs
                        time.sleep(0.33)
                        artist = ytmusic.get_artist(r['artists'][0]['id'])
                        break

        # If we got here and did not match, we might have a common or short 
        # album name like "Gold", or "Blue", or "Greatest Hits"
        # Our next step is to find the artist and then the albums of that artist
        # and look for a good match
        if not matched:
            if not dry_run:
                print(f"No album match. Searching artists for {a['artist']}...")
            # search for the artist by album artist
            time.sleep(0.33)
            artist_search_results = ytmusic.search(a['artist'],"artists")
            artist_match = False
            for r in artist_search_results:
                if r['resultType'] != 'artist':
                    continue
                if artist_match:
                    break

                # check if we have an exact match on the artist name (in lowercase)
                if r['artist'].lower() == a['artist'].lower():
                    artist = r
                    artist_match = True
                    if not dry_run:
                        print(f"{r['artist']} found. browsing albums...")
                    time.sleep(0.33)
                    artist = ytmusic.get_artist(r['browseId'])
                    time.sleep(0.33)
                    # Not every albums array we get back from YT has a browseId and params
                    # It seems the artists with less than 10 or so albums don't have it
                    # If they don't we need to iterate on the albums here
                    if not artist['albums']['browseId'] or not artist['albums']['params']:
                        artist_albums = artist['albums']['results']
                    else:
                        artist_albums = ytmusic.get_artist_albums(artist['albums']['browseId'],artist['albums']['params'], None)
                    for artist_album in artist_albums:
                        # we need to query the music-utils go application for each album
                        # to see if the match is reasonable enough to continue
                        # force an array of artists on this object so that the 
                        # get_album_from_response will work.
                        if not 'artists' in artist_album or len(artist_album['artists']) == 0:
                            artist_album['artists'] = [{}]
                            artist_album['artists'][0]['name'] = r['artist']
                        maybeAlbum = get_album_from_response(artist_album)
                        score, track_match_indexes = cmp_albums(a, maybeAlbum)

                        # If the score is above the threshold, get the entire album info from yt
                        # This threshold is LOWER than before because we know the artist
                        # matches and is to our benefit to look harder for matching tracks
                        # and track comparision does not start until after we cross 
                        # the threshold below.
                        if (score > 0.4):
                            time.sleep(0.33)
                            # get the complete album info by querying with the browseId
                            yt_album = ytmusic.get_album(artist_album['browseId'])
                            # update the score
                            maybeAlbum = get_album_from_response(yt_album)
                            score, track_match_indexes = cmp_albums(a, maybeAlbum, False)

                            # Check if there are alternate versions
                            if 'other_versions' in yt_album:
                                best_other_score, best_other_track_match_indexes, best_other_album = get_best_other_album(ytmusic, yt_album['other_versions'], a)
                                if best_other_score > score:
                                    yt_album = best_other_album
                                    score = best_other_score
                                    track_match_indexes = best_other_track_match_indexes

                            if score > 0.8:
                                matched = True
                                playlistIdsToDownload.append(yt_album['audioPlaylistId'])
                                print(f"very good match on {a['album']}")
                                removed_matched_tracks(a, track_match_indexes)
                                break
                        else:
                            if not dry_run:
                                print(f"score: {score} for {artist_album['title']}")

        # did we match and all tracks on the album are accounted for?
        # This is for when we have no matches on any album
        if not matched:
            if not dry_run:
                print(f"SKIPPED {a['artist']} - {a['album']}")
        
        for t in a['tracks']:
            print(f"missing track: {t['trackNumber']} - {t['title']} - {t['durationSeconds']}")

        # Get all the songs to iterate over to find the stuff in the track_dict
        if len(a['tracks']) > 0:
            num_to_find = len(a['tracks'])
            time.sleep(0.33)
            tracks_browse_result = ytmusic.get_playlist(artist['songs']['browseId'])
            tracks_to_search_over = tracks_browse_result['tracks']
            # sort by title
            tracks_to_search_over.sort(key=return_title)
            cnt = 0
            # iterate over the track to find
            found_tracks = []
            for t in a['tracks']:
                maybe_index = bisect_left(tracks_to_search_over, 
                                          t['title'].lower(), 
                                          key=lambda x: x['title'].lower())
                # score for the a range around the bisect reult but within bounds
                best_t_score = 0.0
                best_track = {}
                # walk from maybe_index down
                i = maybe_index
                t_score = 1.0
                stop = 0.33
                while t_score > stop:
                    try:
                        maybe_track = get_track_from_response(tracks_to_search_over[i])
                    except Exception as e:
                        print(e)
                        print(json.dumps(tracks_to_search_over[maybe_index],indent=2))
                        raise typer.Exit(code=1)
                    t_score = cmp_tracks(t, maybe_track)
                    if t_score > 0.6 and tracks_to_search_over[i]['isExplicit']:
                        t_score += 0.05 # add 5% for being the explicit track
                    if t_score > best_t_score:
                        best_t_score = t_score
                        best_track = tracks_to_search_over[i]
                    i -= 1
                    # check bounds
                    if i < 0:
                        break
                # walk from maybe_index + 1 up
                i = min(maybe_index + 1, len(tracks_to_search_over)-1)
                t_score = 1.0
                while t_score > stop:
                    try:
                        maybe_track = get_track_from_response(tracks_to_search_over[i])
                    except Exception as e:
                        print(e)
                        print(json.dumps(tracks_to_search_over[maybe_index],indent=2))
                        raise typer.Exit(code=1)
                    t_score = cmp_tracks(t, maybe_track)
                    if t_score > 0.6 and tracks_to_search_over[i]['isExplicit']:
                        t_score += 0.05 # add 5% for being the explicit track
                    if t_score > best_t_score:
                        best_t_score = t_score
                        best_track = tracks_to_search_over[i]
                    i += 1
                    if i == len(tracks_to_search_over):
                        break
                if best_t_score > 0.70:
                    print("track found!")
                    print(f"-: {t['title']} - {t['durationSeconds']}")
                    print(f"+: {best_track['title']} - {best_track['duration_seconds']}")
                    # push the track to the download queue
                    videoIdsToDownload.append(best_track['videoId'])
                    # mark to remove the track from the album tracks list
                    found_tracks.append(t)
                    cnt += 1

        # remove all the tracks that we found
        if 'found_tracks' in locals():
            for t in found_tracks:
                a['tracks'].remove(t)

        # if we have any tracks that are still not found, add to skipped
        if len(a['tracks']) > 0:
            skipped.append(a)            
    
    if dry_run:
        report = {}
        report['toDownload'] = playlistIdsToDownload
        report['downloaded'] = albums
        report['skipped'] = skipped
        # print it pretty because why not?
        print(json.dumps(report, indent=2))
        # exit without error
        raise typer.Exit()
    
    # iterate over each yt_playlist and download with yt-dlp
    for pid in playlistIdsToDownload:
        ret_val = subprocess.call(f"yt-dlp --cookies {cookies_txt} -P {output_dir} -o \"%(album)s/%(autonumber)02d - %(track)s.%(ext)s\" -x --audio-format mp3 --add-metadata \"https://music.youtube.com/playlist?list={pid}\"", shell=True)

    for vid in videoIdsToDownload:
        ret_val = subprocess.call(f"yt-dlp --cookies {cookies_txt} -P {output_dir} -o \"%(album)s/%(autonumber)02d - %(track)s.%(ext)s\" -x --audio-format mp3 --add-metadata \"https://music.youtube.com/watch?v={vid}\"", shell=True)

def cmp_albums(album1, album2, ignore_tracks=True, verbose=False, threshold=0.5):
    print(f"cmp_albums: {album1['album']} to {album2['album']}...")
    # create the json request 
    # {
    #   "album1": {...}, 
    #   "album2": {...}, 
    #   "threshold": 0.5
    # }
    request = {}
    request['album1'] = album1
    request['album2'] = album2
    request['compareType'] = 'tracks'
    if ignore_tracks:
        request['compareType'] = 'ignoretracks'
    request['threshold'] = threshold
    if verbose:
        print(json.dumps(request,indent=2))
    # serialize to JSON bytes
    json_bytes = json.dumps(request).encode('utf-8')
    # Base64 encode the bytes
    base64_bytes = base64.b64encode(json_bytes)
    # Bytes to Base64 string
    base64_message = base64_bytes.decode('ascii')
    # call music-utils
    result = subprocess.run(["music-utils","compare","albums","--data", base64_message], capture_output=True, text=True)
    if (result.returncode != 0):
        # fail out if the command returns an error
        print(json.dumps(request, indent=2))
        print("music-utils FAILED. Request above")
        raise typer.Exit(code=1)
    # serialize the json
    data = json.loads(result.stdout)
    #if data['score'] > threshold:
    #    print(json.dumps(data, indent=2))

    return data['score'], data['trackMatchIndexes']

def get_album_from_response(yt_album):
    album = {}
    album['album'] = yt_album['title']
    album['artist'] = ''
    # iterate over the album artists
    for artist in yt_album['artists']:
        album['artist'] += artist['name']
        album['artist'] += ' '
    # remove leading and trailing spaces
    album['artist'] = album['artist'].strip()
    if 'trackCount' in yt_album:
        album['totalTracks'] = yt_album['trackCount']
    album['tracks'] = []
    if 'tracks' in yt_album:
        for yt_track in yt_album['tracks']:
            track = get_track_from_response(yt_track, album['artist'], album['totalTracks'])
            album['tracks'].append(track)
    return album

def get_track_from_response(yt_track, album_artist='', total_tracks=0):
    track = {}
    track['title'] = yt_track['title']
    track['artist'] = ''
    for artist in yt_track['artists']:
        track['artist'] += artist['name']
        track['artist'] += ' '
    # remove leading and trailing spaces
    track['artist'] = track['artist'].strip()
    if 'trackNumber' in yt_track:
    	track['trackNumber'] = yt_track['trackNumber']
    track['totalTracks'] = total_tracks
    if 'name' in yt_track['album']:
        track['album'] = yt_track['album']['name']
    else:
        track['album'] = yt_track['album']
    track['albumArtist'] = album_artist
    track['durationSeconds'] = yt_track['duration_seconds']
    return track

def cmp_tracks(track1, track2, type='track', threshold=0.7):
    #print(f"cmp_tracks: {track1['title']} to {track2['title']}...")
    # create the json request 
    # {
    #   "track1": {...},
    #   "track2": {...},
    #   "compareType": "track", 
    #   "threshold": 0.8
    # }
    request = {}
    request['track1'] = track1
    request['track2'] = track2
    request['compareType'] = type
    request['threshold'] = threshold
    # serialize to JSON bytes
    json_bytes = json.dumps(request).encode('utf-8')
    # Base64 encode the bytes
    base64_bytes = base64.b64encode(json_bytes)
    # Bytes to Base64 string
    base64_message = base64_bytes.decode('ascii')
    # call music-utils
    result = subprocess.run(["music-utils","compare","tracks","--data", base64_message], capture_output=True, text=True)
    if (result.returncode != 0):
        # fail out if the command returns an error
        raise typer.Exit(code=1)
    # serialize the json
    data = json.loads(result.stdout)
    #if data['score'] > threshold:
    #    print(json.dumps(data, indent=2))

    return data['score']

def get_best_other_album(ytmusic, other_versions, target):
    best_score = 0.0
    best_album = ''
    best_track_match_indexes = {}
    for a in other_versions:
        # get the album
        time.sleep(0.33)
        album = ytmusic.get_album(a['browseId'])
        # score it
        maybeAlbum = get_album_from_response(album)
        score, track_match_indexes = cmp_albums(target, maybeAlbum, False)
        # 5% bonus for being explicit
        if a['isExplicit']:
            score = score + 0.05
        # see if it is the best
        if score > best_score:
            best_score = score
            best_album = album
            best_track_match_indexes = track_match_indexes
    return best_score, best_track_match_indexes, best_album

def removed_matched_tracks(a, track_match_indexes):
    to_delete = []
    # create a lit of int from the keys
    for i in track_match_indexes:
        to_delete.append(int(i))
    # sort so the list goes from biggest to smallest
    to_delete.sort(reverse=True)
    # iterate over the list, deleting the largest index to the smallest
    for i in to_delete:
        del a['tracks'][i]

def return_title(track):
    return track['title']
