# Markdown View

[![Snap Status](https://build.snapcraft.io/badge/mapitman/mdview.svg)](https://build.snapcraft.io/user/mapitman/mdview)

Formats markdown and launches it in a browser.

## Install

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/mdview)

On Linux, you can install [mdview](https://snapcraft.io/mdview) from the snap store:

```sh
sudo snap install mdview
```

Don't have snapd? [Get set up for snaps](https://docs.snapcraft.io/core/install).

If you have Golang installed...
```sh
go get github.com/mapitman/mdview
```

Don't have Golang? [Get it now](https://golang.org/doc/install).

Otherwise, grab the correct binary [here](https://github.com/mapitman/mdview/releases/).


## Usage

```text
Usage:
mdview [options] <filename>
Formats markdown and launches it in a browser.
  -b    Bare HTML with no style applied.
  -bare
        Bare HTML with no style applied.
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
