# Release Checklist

Use this checklist when preparing a new release.

## Pre-Release

- [ ] Ensure all code changes are committed
- [ ] Update version in code (if needed)
- [ ] Update CHANGELOG.md with release notes
- [ ] Review changes: `git log --oneline [previous-tag]..HEAD`
- [ ] All tests passing locally
- [ ] Test local builds (optional):
  ```bash
  make rpm-local VERSION=1.0.0
  make deb VERSION=1.0.0
  ```

## Build Verification (Optional)

Test packages locally before releasing:

- [ ] RPM installs without errors:
  ```bash
  sudo dnf install ./dist/mdview-*.rpm
  ```
- [ ] DEB installs without errors:
  ```bash
  sudo dpkg -i ./dist/mdview-*.deb
  ```
- [ ] Binary works:
  ```bash
  which mdview
  mdview --version
  mdview --help
  ```
- [ ] Man page is accessible:
  ```bash
  man mdview
  ```

## Create Release Tag

- [ ] Commit all changes
- [ ] Create tag:
  ```bash
  git tag -a v1.0.0 -m "Release v1.0.0"
  ```
- [ ] Verify tag:
  ```bash
  git show v1.0.0
  ```
- [ ] Push tag:
  ```bash
  git push origin v1.0.0
  ```

## GitHub Actions Verification

- [ ] Go to Actions tab
- [ ] Check "Release" workflow:
  - [ ] build-binaries job completed
  - [ ] build-rpm job completed
  - [ ] release job completed
- [ ] All jobs passed (green checkmarks)
- [ ] No warnings or errors in logs

## Release Page Verification

- [ ] Go to Releases page
- [ ] Release created with correct version
- [ ] Check release contains:
  - [ ] Binary distributions for all platforms (tar.gz, zip)
  - [ ] RPM packages (.x86_64.rpm, .src.rpm)
  - [ ] Debian packages (.deb files for amd64, arm64, i386)
  - [ ] Release notes/description
- [ ] Test download and install:
  ```bash
  # For RPM
  wget https://github.com/mapitman/mdview/releases/download/v1.0.0/mdview-1.0.0-1.fc*.x86_64.rpm
  sudo dnf install mdview-1.0.0-1.fc*.x86_64.rpm
  
  # For DEB
  wget https://github.com/mapitman/mdview/releases/download/v1.0.0/mdview_1.0.0_amd64.deb
  sudo dpkg -i mdview_1.0.0_amd64.deb
  
  # Verify
  mdview --version
  ```

## Post-Release

- [ ] Uninstall test version (if installed)
- [ ] Announce release (if applicable)
- [ ] Update README with new version info (if needed)
- [ ] Tag appears in GitHub releases list
- [ ] Local cleanup: `make rpm-clean`

## Emergency Procedures

### If workflow fails:
1. Check GitHub Actions logs for errors
2. Fix issues locally using `make ci-sim` or package-specific builds
3. Delete tag: `git tag -d v1.0.0`
4. Delete remote tag: `git push origin --delete v1.0.0`
5. Fix issues and retry

### If release is incorrect:
1. Delete release from GitHub releases page
2. Delete tag: `git tag -d v1.0.0 && git push origin --delete v1.0.0`
3. Fix issues locally
4. Create new tag and push again

### If package build fails but other builds succeed:
1. Check specific job logs in GitHub Actions
2. Test locally with CI simulation: `make ci-sim`
3. Fix issues and create new tag

## Version Format Notes

- Use semantic versioning: `v1.0.0`, `v1.0.1`, `v1.1.0`, etc.
- Git tags include `v` prefix
- Makefile commands use no prefix: `make rpm VERSION=1.0.0`
- RPM packages appear as `1.0.0-1.fc*.x86_64.rpm`
- DEB packages appear as `mdview_1.0.0_amd64.deb`

## Rollback Procedure

If released version has critical bugs:
1. Fix issues in code
2. Create new tag for patch:
   ```bash
   git tag v1.0.1
   git push origin v1.0.1
   ```
3. Document issue in old release notes

## Notes

- All builds happen automatically in GitHub Actions on tag push
- Packages are built in containers for consistency
- Multiple output formats included in each release
- Consider using pre-release flag for testing
- See `PACKAGING.md` for detailed build instructions

---

**Last Updated**: December 2024
**For**: mdview project
**Maintainer**: mapitman
