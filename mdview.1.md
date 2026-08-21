% MDVIEW(1) Version 1.4.0 | "Markdown View" Documentation

# NAME

**mdview** — Formats markdown and launches it in a browser.

# SYNOPSIS

**mdview** _filename_  
**mdview** **-**  
**mdview** \[**-h**|**--help**|**-v**|**--version**]

# DESCRIPTION

Formats a markdown file as HTML, writes it to a temporary file and
then launches that file in the default web browser. By default, it will
use the operating system's default temporary directory unless the
environment variable MDVIEW_DIR is set.

If _filename_ is a single dash (**-**), the markdown is read from
standard input. Relative image paths are then resolved against the
current working directory.

## Options

**-b**, **-bare**

Bare HTML with no style applied.

**-h**, **-help**

Prints mdview help message.

**-o** _filename_

Output filename. (Optional)

**-v**, **-version**

Prints mdview version.

# BUGS

See GitHub Issues: <https://github.com/mapitman/mdview/issues>

# AUTHOR

Mark Pitman <mark@pitman.io> <https://pitman.io>
