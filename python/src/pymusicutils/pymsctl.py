import typer
from typing_extensions import Annotated
import os
from .validate import fileOrDirectoryExists
import json
import subprocess

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
    albums = []
    # create a tuple for skipped albums (no good matches etc)
    skipped = []

    # iterate over the albums
    for a in exif_report["albums"]:
        continue
        #print(f"album: {a["album"]}")
        #print(f"artist: {a["artist"]}")
        #print(f"totalTracks: {a["totalTracks"]}")
        #print()
        # TODO on each album, search with ytmusicapi

    if dry_run:
        report = {}
        report['downloads'] = albums
        report['skipped'] = skipped
        # print it pretty because why not?
        print(json.dumps(report, indent=2))
        # exit without error
        raise typer.Exit()
    
    # iterate over each yt_playlist and download with yt-dlp
    ret_val = subprocess.call("yt-dlp --help", shell=True)


