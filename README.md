# Markdown View

[![Build](https://github.com/mapitman/mdview/actions/workflows/build.yml/badge.svg)](https://github.com/mapitman/mdview/actions/workflows/build.yml)

Formats markdown and launches it in a browser.

## Usage

By default, `mdview` tries to use your operating system's temporary
directory to write HTML files to. If that doesn't work for you, you can
set an environment variable that it will use instead. For example, on
Ubuntu Linux, Firefox is packaged as a Snap and is unable to read from
`/tmp`. I get around this by setting `MDVIEW_DIR` like so:

```sh
export MDVIEW_DIR=$HOME/mdview-temp
```


```text
Usage:
mdview [options] <filename>
Formats markdown and launches it in a browser.
If the environment variable MDVIEW_DIR is set, the temporary file will be written there.
  -b Bare HTML with no style applied.
  -bare
     Bare HTML with no style applied.
  -h Prints mdview help message.
  -help
     Prints mdview help message.
  -o string
     Output filename. (Optional)
  -v Prints mdview version.
  -version
     Prints mdview version.

```

If you do not supply an output file, mdview will write a file to your
operating system's default temp directory or to the value of MDVIEW_DIR.

The generated HTML will conform to your system's light or dark theme
setting, as long as your browser supports that feature.

### Thanks

Thanks to [sindresorhus](https://github.com/sindresorhus/github-markdown-css) for the GitHub style css.

## Installation

### Arch Linux (and derivatives)

Markdown View is now available in the [AUR](https://aur.archlinux.org/packages/mdview/)
If you have an AUR helper like `yay`, installing is as easy as:
```
yay -S mdview
```

### Debian Package

If you're running Debian or a derivative like Ubuntu or Pop!_OS, you can
use [deb-get](https://github.com/wimpysworld/deb-get) to install mdview.

```sh
deb-get install mdview
```

If you don't want to use `deb-get`, you can download the package and
manually install it from the
[Releases](https://github.com/mapitman/mdview/releases) page.

```properties
curl -s https://api.github.com/repos/mapitman/mdview/releases/latest \
| grep "browser_download_url.*amd64.deb" \
| cut -d '"' -f 4 \
| xargs curl -L -o mdview_last_amd64.deb
sudo dpkg --install mdview_last_amd64.deb
```

To remove the package:

```sh
sudo dpkg --remove mdview
```

### Snap Package

_Update: The snap package has been fixed and the latest version is now available as a snap._ 🥳

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

Don't have snapd?  
[Get set up for snaps](https://docs.snapcraft.io/core/install).

### Manual Download and Install

Grab the correct binary for your operating system
[here](https://github.com/mapitman/mdview/releases/).

### Compile Yourself

If you have Golang installed...
```sh
go get github.com/mapitman/mdview
```

Don't have Golang? [Get it now](https://golang.org/doc/install).


