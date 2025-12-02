# RPM Release Checklist

Use this checklist when preparing a new release with RPM packages.

## Pre-Release

- [ ] Ensure all code changes are committed
- [ ] Update version in code (if needed)
- [ ] Update CHANGELOG.md with release notes
- [ ] Review changes: `git log --oneline [previous-tag]..HEAD`
- [ ] All tests passing locally
- [ ] Test local build:
  ```bash
  ./build-rpm.sh build 1.0.0
  sudo dnf install ./dist/mdview-*.rpm
  mdview --version
  ```

## Build Verification

- [ ] Local build successful:
  ```bash
  ./build-rpm.sh local 1.0.0
  ```
- [ ] RPM installs without errors:
  ```bash
  sudo dnf install ./dist/mdview-*.rpm
  ```
- [ ] Binary works:
  ```bash
  which mdview
  mdview --version
  ```
- [ ] Man page is accessible:
  ```bash
  man mdview
  ```
- [ ] Can test with help:
  ```bash
  mdview --help
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
- [ ] Release v1.0.0 created
- [ ] Check release contains:
  - [ ] Binary distributions (tar.gz, zip)
  - [ ] RPM packages (.x86_64.rpm, .src.rpm)
  - [ ] Debian packages (.deb files)
  - [ ] Release notes/description
- [ ] Downloads work for RPM:
  ```bash
  wget https://github.com/mapitman/mdview/releases/download/v1.0.0/mdview-1.0.0-1.fc*.x86_64.rpm
  sudo dnf install mdview-1.0.0-1.fc*.x86_64.rpm
  mdview --version
  ```

## Post-Release

- [ ] Uninstall test version:
  ```bash
  sudo dnf remove mdview
  ```
- [ ] Announce release (if applicable)
- [ ] Update README with new version info
- [ ] Consider submitting to COPR if using for community
- [ ] Tag appears in GitHub releases list
- [ ] Local cleanup:
  ```bash
  ./build-rpm.sh clean
  ```

## RPM-Specific Checks

- [ ] Binary RPM installs dependencies correctly
- [ ] Source RPM (.src.rpm) exists
- [ ] RPM includes man page
- [ ] RPM includes documentation
- [ ] RPM includes license file
- [ ] Post-install tests work:
  ```bash
  mdview --version
  mdview --help
  man mdview
  ```

## Emergency Procedures

### If workflow fails:
1. Check GitHub Actions logs for errors
2. Fix issues locally: `./build-rpm.sh build 1.0.0`
3. Delete tag: `git tag -d v1.0.0`
4. Delete remote tag: `git push origin --delete v1.0.0`
5. Fix issues and retry

### If release is incorrect:
1. Delete release from GitHub releases page
2. Delete tag: `git tag -d v1.0.0 && git push origin --delete v1.0.0`
3. Fix issues locally
4. Create new tag and push again

### If RPM build fails but other builds succeed:
1. Manual RPM build:
   ```bash
   ./build-rpm.sh build 1.0.0
   ```
2. Upload to release manually if needed

## Version Format Notes

- Use semantic versioning: `v1.0.0`, `v1.0.1`, `v1.1.0`, etc.
- RPM automatically converts to `1.0.0-1.fc*` format
- Version in tag includes `v` prefix
- Version in make command uses no prefix

## Distribution

### GitHub Releases
- ✅ Automatic via push tag
- All packages included
- Directly downloadable

### Fedora COPR
- [ ] Create account if needed
- [ ] Create project
- [ ] Link GitHub repo
- [ ] Enable Fedora versions
- [ ] Wait for builds to complete
- [ ] Announce availability

### Direct Installation
Users can install directly:
```bash
sudo dnf install https://github.com/mapitman/mdview/releases/download/v1.0.0/mdview-1.0.0-1.fc*.x86_64.rpm
```

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

- All builds happen automatically in GitHub Actions
- RPMs are built in Fedora container for consistency
- Multiple output formats included in release
- Consider using pre-release for testing
- First-time setup may take extra review by GitHub

---

**Last Updated**: November 2025
**For**: mdview project
**Maintainer**: mapitman
