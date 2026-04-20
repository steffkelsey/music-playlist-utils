import typer
from typing_extensions import Annotated
import os
from .validate import fileOrDirectoryExists
import json
import time
import subprocess
from ytmusicapi import YTMusic

def pymsctl(
    input_file: Annotated[str, typer.Option(help="The json report with albums to search for")],
    output_dir: Annotated[str, typer.Option(help="The directory to size downloaded msuic files")],
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
        # TODO is there a fuzzy match on the album title?
        # TODO is there a fuzzy match with the album artists?
        #print(f"album: {a["album"]}")
        #print(f"artist: {a["artist"]}")
        #print(f"totalTracks: {a["totalTracks"]}")
        #print()
        # on each album, search with ytmusicapi
        album_search_results = ytmusic.search(a["album"], 'albums')
        for r in album_search_results:
            if r['resultType'] != 'album':
                continue
            # does the title match exactly?
            if r['title'].lower() != a['album'].lower():
                p = percentWordMatch(r['title'].lower(), a['album'].lower())
                if p < 0.6:
                    print(f"title does not match '{r['artists'][0]['name']}' - '{r['title']}'")
                    continue
            # get the complete album info by querying with the browseId
            yt_album = ytmusic.get_album(r['browseId'])
            # check that the total tracks match
            if a['totalTracks'] == str(yt_album['trackCount']):
                # check that the artist matches for the album

                matched = True
                print(f"MATCHED: {yt_album['title']}")
                break
            else:
                print(f"total tracks did not match: {yt_album['trackCount']}, '{yt_album['artists'][0]['name']}' - '{yt_album['title']}' != {a['totalTracks']}")
                continue
                
            playlistIdsToDownload.append(yt_album['audioPlaylistId'])
        if not matched:
            skipped.append(a)
            if not dry_run:
                print(f"SKIPPED {a['artist']} - {a['album']}")
        # pause to prevent 429 request overload error from Youtube API
        time.sleep(1)

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
    #ret_val = subprocess.call("yt-dlp --help", shell=True)

def percentWordMatch(p1, p2):
    p1Words = p1.split()
    p1MeaningfulWordMap = {}
    for w in p1Words:
        if len(w) > 0:
            p1MeaningfulWordMap[w] = True
    p2Words = p2.split()
    matches = 0
    for w in p2Words:
        if len(w) > 0 and p1MeaningfulWordMap.get(w):
            matches += 1
    return matches/len(p1MeaningfulWordMap)
