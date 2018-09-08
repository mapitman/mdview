# Markdown View

[![Snap Status](https://build.snapcraft.io/badge/mapitman/mdview.svg)](https://build.snapcraft.io/user/mapitman/mdview)

Formats markdown and launches it in a browser.

## Install

On Linux, you can install [mdview](https://snapcraft.io/mdview) from the snap store:

```sh
sudo snap install mdview
```

If you have Go installed...
```sh
go get github.com/mapitman/mdview
```

Otherwise, grab the correct binary [here](https://github.com/mapitman/mdview/releases/tag/1.0.0).


## Usage

```text
mdview [options] <filename>
Formats markdown and launches it in a browser.
  -h    Prints mdview help message.
  -help
        Prints mdview help message.
  -o string
        Output filename. (Optional)
  -v    Prints mdview version.
  -version
        Prints mdview version.
```

If you do not supply an output file, mdview will write a file to your 
operating system's default temp directory.
