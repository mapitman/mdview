
1.8.0 / 2026-01-18
==================

  * Add self-contained Mermaid diagram support via Goldmark (#49)
  * Add Copilot instructions for repository (#51)

1.7.0 / 2025-12-20
==================

  * Convert relative image links to base64 data URIs (markdown and HTML) (#47)

1.6.4 / 2024-12-31
==================

  * Merge pull request #39 from mapitman/bug/handle-file-not-found
  * Handle file not found error
  * Revert "Update snap build to set version"

1.6.3 / 2024-11-30
==================

  * Merge pull request #33 from mapitman/handle-permission-denied
  * Add vscode settings
  * Update snap build to set version
  * Handle permission denied error a little better
  * 1 version update (readme) (#31)

1.6.2 / 2024-03-23
==================

  * Merge pull request #28 from mapitman/dependabot/go_modules/golang.org/x/text-0.3.8
  * Bump golang.org/x/text from 0.3.2 to 0.3.8
  * Update README.md

1.6.1 / 2024-02-19
==================

  * Merge pull request #27 from mapitman/handle-inline-html
  * Fix build
  * Handle in-line HTML

1.6.0 / 2024-02-19
==================

  * Merge pull request #25 from mapitman/update-readme
  * Update README
  * Merge pull request #24 from mapitman/dependabot/go_modules/golang.org/x/text-0.3.8
  * Bump golang.org/x/text from 0.3.2 to 0.3.8
  * Merge pull request #23 from mapitman/update-actions
  * Update to latest versions of some github actions
  * Merge pull request #22 from mapitman/dark-mode
  * Update to support light/dark themes via OS setting
  * Merge pull request #21 from mapitman/update-readme
  * Update docs and Snap build process
  * Add build status badge
  * Auto-created releases will be drafts

1.5.0 / 2023-09-17
==================

  * Prepare for next release
  * Merge pull request #15 from mapitman/suppress-browser-messages
  * Suppress stderr and stdout messages from browser
  * Merge pull request #13 from mapitman/support-env-var-for-output
  * Support reading an environment variable for the directory to write to
  * Merge pull request #11 from mapitman/add-actions
  * Add automated build and release actions

1.4.1 / 2021-12-28
==================

  * Fix build with go 1.17
  * Add info about AUR to README

1.4.0 / 2020-11-08
==================

  * Add manpage to Linux tarballs
  * Add files and config to build Debian package
  * Write temp files into $HOME/mdview-temp when installed via snap
  * add snap store image and link

1.3.0 / 2018-11-16
==================

  * Update readme and history
  * Add ability to render markdown file with no style applied
  * Update Makefile to use version variable
  * Update link to releases
  * Update README with links to install snapd and golang
  * trim history file
  * Add changelog
  * Add vscode-specific files
  * Bump version for next release
  * Extract HTML title from markdown

1.2.0 / 2018-11-02
==================

  * Bump app version
  * Bump snap version
  * Merge pull request #5 from eaglersdeveloper/opening-from-file-manager
  * Add more information to desktop file
  * Bump snap version
  * update gitignore

1.1.0 / 2018-10-30
==================

  * Update version to 1.1
  * Merge pull request #4 from eaglersdeveloper/opening-from-file-manager
  * Add the ability to open from file manager
  * Merge pull request #3 from evandandrea/fix-snap
  * Fix snapcraft.yaml (use go1.8)
  * update README with link to snap store
  * Fix desktop file
  * Add desktop file to get rid of warning when snap is built
  * Add parts/ to gitignore
  * Add unity7 plug so xdg-open works in snap
  * candidate is not the correct value, should be stable
  * switch snap to candidate
  * Add snap badge
  * Update so that classic confinement is not required for snap
  * Fix make install
  * Add gopath to go build line
  * Fix gopath setting in makefile
  * Add gopath
  * Change confinement to classic
  * update makefile to have a snap task
  * Add go get to makefile
  * Add golang as build dependency
  * Update snap yaml to build the right thing
  * Update snapcraft.yaml
  * Create binary archive files as part of build
  * Update README
  * Initial commit

1.0.0 / 2018-08-25
==================

  * Initial commit
