#!/bin/bash
# mdview RPM build helper script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_usage() {
    cat << EOF
${BLUE}mdview RPM Build Helper${NC}

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  build [VERSION]      Build RPM (default: latest git tag)
  local [VERSION]      Build and copy to dist/ (default: latest git tag)
  setup                Initialize RPM build environment
  clean                Clean all RPM build artifacts
  info                 Show build information
  help                 Show this help message

Examples:
  $0 build                    # Build using latest git tag
  $0 build 1.0.0             # Build specific version
  $0 local 1.0.0             # Build and copy to dist/
  $0 setup                   # Set up build environment
  $0 clean                   # Clean build artifacts
  $0 info                    # Show environment info

EOF
}

check_dependencies() {
    # Map of command names to package names for Fedora
    declare -A deps=(
        ["rpmbuild"]="rpm-build"
        ["go"]="golang"
        ["pandoc"]="pandoc"
        ["make"]="make"
        ["git"]="git"
    )
    local missing_cmds=()
    local missing_pkgs=()
    
    for cmd in "${!deps[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            missing_cmds+=("$cmd")
            missing_pkgs+=("${deps[$cmd]}")
        fi
    done
    
    if [ ${#missing_cmds[@]} -ne 0 ]; then
        echo -e "${RED}Error: Missing required commands: ${missing_cmds[*]}${NC}"
        echo
        echo -e "${YELLOW}These commands are required but not found in your PATH.${NC}"
        echo
        echo -e "${YELLOW}To install on Fedora:${NC}"
        echo "  sudo dnf install -y ${missing_pkgs[*]}"
        echo
        echo -e "${YELLOW}If packages are already installed, ensure the commands are in your PATH.${NC}"
        exit 1
    fi
}

get_version() {
    local provided_version="$1"
    
    if [ -n "$provided_version" ]; then
        # Remove leading 'v' if present
        echo "${provided_version#v}"
    else
        # Try to get from git tag
        local tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
        if [ -n "$tag" ]; then
            echo "${tag#v}"
        else
            echo "dev"
        fi
    fi
}

show_info() {
    echo -e "${BLUE}mdview RPM Build Information${NC}"
    echo
    
    check_dependencies
    
    echo -e "${GREEN}✓ Dependencies: OK${NC}"
    echo
    
    echo "Environment:"
    echo "  RPM Build Dir: $HOME/rpmbuild/"
    echo "  .rpmmacros: $HOME/.rpmmacros"
    echo
    
    # Try to get version
    local version=$(get_version "")
    echo "Version Information:"
    echo "  Current version: $version"
    echo
    
    echo "Installed packages:"
    rpmbuild --version 2>/dev/null || echo "  rpmbuild: not found"
    go version 2>/dev/null || echo "  go: not found"
    pandoc --version 2>/dev/null | head -1 || echo "  pandoc: not found"
    make --version 2>/dev/null | head -1 || echo "  make: not found"
}

build_rpm() {
    local version="$1"
    
    check_dependencies
    
    echo -e "${BLUE}Building mdview RPM${NC}"
    echo "Version: $version"
    echo
    
    # Run make
    if make rpm VERSION="$version"; then
        echo
        echo -e "${GREEN}✓ RPM build successful!${NC}"
        echo
        echo "Packages created in:"
        echo "  Binary RPM: $HOME/rpmbuild/RPMS/x86_64/"
        echo "  Source RPM: $HOME/rpmbuild/SRPMS/"
        echo
        
        # Show what was built
        if [ -d "$HOME/rpmbuild/RPMS/x86_64" ]; then
            echo "Built packages:"
            ls -lh "$HOME/rpmbuild/RPMS/x86_64/mdview-"* 2>/dev/null || echo "  (none found)"
        fi
        if [ -d "$HOME/rpmbuild/SRPMS" ]; then
            ls -lh "$HOME/rpmbuild/SRPMS/mdview-"* 2>/dev/null || true
        fi
        echo
        echo -e "${YELLOW}To install:${NC}"
        echo "  sudo dnf install $HOME/rpmbuild/RPMS/x86_64/mdview-*.rpm"
    else
        echo -e "${RED}✗ RPM build failed!${NC}"
        exit 1
    fi
}

build_local() {
    local version="$1"
    
    check_dependencies
    
    echo -e "${BLUE}Building mdview RPM (local mode)${NC}"
    echo "Version: $version"
    echo
    
    # Run make
    if make rpm-local VERSION="$version"; then
        echo
        echo -e "${GREEN}✓ RPM build and copy successful!${NC}"
        echo
        
        # Show what was copied
        if [ -d "dist" ]; then
            echo "Packages in dist/:"
            ls -lh dist/mdview-* 2>/dev/null || echo "  (none found)"
        fi
        echo
        echo -e "${YELLOW}To install:${NC}"
        echo "  sudo dnf install ./dist/mdview-*.rpm"
    else
        echo -e "${RED}✗ RPM build failed!${NC}"
        exit 1
    fi
}

setup_environment() {
    echo -e "${BLUE}Setting up RPM build environment${NC}"
    
    if make rpm-setup; then
        echo -e "${GREEN}✓ RPM build environment setup complete!${NC}"
        echo
        echo "Directory structure created:"
        echo "  $HOME/rpmbuild/BUILD"
        echo "  $HOME/rpmbuild/RPMS"
        echo "  $HOME/rpmbuild/SOURCES"
        echo "  $HOME/rpmbuild/SPECS"
        echo "  $HOME/rpmbuild/SRPMS"
    else
        echo -e "${RED}✗ Setup failed!${NC}"
        exit 1
    fi
}

clean_artifacts() {
    echo -e "${BLUE}Cleaning RPM build artifacts${NC}"
    
    if [ -d "$HOME/rpmbuild" ]; then
        echo "Removing: $HOME/rpmbuild"
        rm -rf "$HOME/rpmbuild"
        echo -e "${GREEN}✓ Removed rpmbuild directory${NC}"
    fi
    
    if [ -d "dist" ]; then
        echo "Removing: dist/mdview-*.rpm"
        rm -f dist/mdview-*.rpm
        if [ -z "$(ls -A dist/ 2>/dev/null)" ]; then
            rmdir dist 2>/dev/null || true
        fi
        echo -e "${GREEN}✓ Removed local dist files${NC}"
    fi
    
    echo
    echo -e "${GREEN}✓ Cleanup complete!${NC}"
}

# Main
main() {
    local command="${1:-help}"
    local arg="${2:-}"
    
    case "$command" in
        build)
            build_rpm "$(get_version "$arg")"
            ;;
        local)
            build_local "$(get_version "$arg")"
            ;;
        setup)
            setup_environment
            ;;
        clean)
            clean_artifacts
            ;;
        info)
            show_info
            ;;
        help)
            print_usage
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"
