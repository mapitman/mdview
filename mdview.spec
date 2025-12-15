%define debug_package %{nil}

Name:           mdview
Version:        %{_version}
Release:        1%{?dist}
Summary:        Formats markdown and launches it in a browser

License:        MIT
URL:            https://github.com/mapitman/mdview
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  pandoc

Requires:       xdg-utils

%description
Markdown View is a utility that formats markdown and launches it in your 
default browser. It supports custom styling and can write output to a file.

%prep
%setup -q

%build
# Build the binary with version information
%{__make} VERSION=%{version} bin/linux-amd64/mdview

%install
# Install the binary
%{__install} -Dp -m 0755 bin/linux-amd64/mdview %{buildroot}%{_bindir}/mdview

# Install the man page
%{__install} -Dp -m 0644 mdview.1 %{buildroot}%{_mandir}/man1/mdview.1

%files
%doc README.md CHANGELOG.md
%license LICENSE
%{_bindir}/mdview
%{_mandir}/man1/mdview.1*

%changelog
* %(date +"%a %b %d %Y") Mark Pitman <mark@mapitman.com> - %{version}-1
- Initial RPM package
