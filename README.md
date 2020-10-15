# Markdown View

[![Snap Status](https://build.snapcraft.io/badge/mapitman/mdview.svg)](https://build.snapcraft.io/user/mapitman/mdview)

Formats markdown and launches it in a browser.

## Installation

### Debian Package

If you're running Debian or a derivative like Ubuntu or Pop!_OS,
download the [deb package](https://github.com/mapitman/mdview/releases/download/1.4.0/mdview-1.4.0_amd64.deb).

```sh
curl -O https://github.com/mapitman/mdview/releases/download/1.4.0/mdview-1.4.0_amd64.deb
sudo dpkg --install mdview-1.4.0_amd64.deb
```

To remove the package:

```sh
sudo dpkg --remove mdview
```

### Snap Package

On Linux, you can install [mdview](https://snapcraft.io/mdview) from the snap store. This option is only viable if the files
you want to view are in your home directory. If you need to view
files in other locations, try an alternate installation method.

_Note_: A side effect of the sandboxing of Snap packages is that every time
`mdview` is executed, Snap will prompt to allow writing
the temporary file. If that is not acceptable, please choose an
alternate installation method.

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/mdview)

```sh
sudo snap install mdview
```

Don't have snapd? [Get set up for snaps](https://docs.snapcraft.io/core/install).

### Manual Download and Install

Grab the correct binary for your operating system
[here](https://github.com/mapitman/mdview/releases/).

### Compile Yourself

If you have Golang installed...
```sh
go get github.com/mapitman/mdview
```

Don't have Golang? [Get it now](https://golang.org/doc/install).

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
