from pathlib import Path

def fileOrDirectoryExists(path):
    return Path.exists(path)

