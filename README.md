# Markdown View

[![Snap Status](https://build.snapcraft.io/badge/mapitman/mdview.svg)](https://build.snapcraft.io/user/mapitman/mdview)

Formats markdown and launches it in a browser.

## Installation

### Arch Linux (and derivatives)

Markdown View is now available in the [AUR](https://aur.archlinux.org/packages/mdview/)
If you have an AUR helper like `yay`, installing is as easy as:
```
yay -S mdview
```

### Debian Package

If you're running Debian or a derivative like Ubuntu or Pop!_OS,
download the deb package for the release you'd like to install form the
[Releases](https://github.com/mapitman/mdview/releases) page.
Install with `dpkg -i`

```sh
curl -O https://github.com/mapitman/mdview/releases/download/1.4.0/mdview-1.4.0_amd64.deb
sudo dpkg --install mdview-1.4.0_amd64.deb
```

To remove the package:

```sh
sudo dpkg --remove mdview
```

### Snap Package

_Update 2023-09-17_: I can't really recommend this method of installation
as it is cumbersome. Also, the Snap build seems to be broken and I don't
really know what to do to fix it!

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

By default, `mdview` tries to use your operating system's temporary
directory. To write HTML files to. If that doesn't work for you, you can
set an environment variable that it will use instead. For example, on
Ubuntu Linux, Firefox is packaged as a Snap and is unable to read from
`/tmp`. I get around this by setting `MDVIEW_DIR` like so:

```sh
export MDVIEW_DIR=$HOME/snap/firefox/mdview
```


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
operating system's default temp directory or to the value of MDVIEW_DIR.
