%global debug_package %{nil}

Name:           system-hub
Version:        0.1.0
Release:        1%{?dist}
Summary:        System administration hub for Linux

License:        MIT
URL:            https://github.com/Lucky2356/system-hub
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.26
BuildRequires:  gcc
BuildRequires:  pkgconfig
# Fyne (GLFW/OpenGL) build-time headers
BuildRequires:  mesa-libGL-devel
BuildRequires:  libX11-devel
BuildRequires:  libXrandr-devel
BuildRequires:  libXcursor-devel
BuildRequires:  libXi-devel
BuildRequires:  libXinerama-devel
BuildRequires:  libXxf86vm-devel
BuildRequires:  libxkbcommon-devel

Requires:       systemd
Recommends:     docker
Recommends:     lm_sensors

%description
System Hub is a desktop application for monitoring system state,
managing systemd services, Docker containers and logs.

%prep
%autosetup

%build
# Offline build: vendor directory is expected in the source tarball
# (run `go mod vendor` before `make dist`). Falls back to module mode if absent.
export CGO_ENABLED=1
go build -trimpath \
    -ldflags="-s -w -X github.com/Lucky2356/system-hub/internal/version.Version=%{version}" \
    -o system-hub ./cmd/system-hub

%install
install -Dpm0755 system-hub %{buildroot}%{_bindir}/system-hub
install -Dpm0644 packaging/system-hub.desktop %{buildroot}%{_datadir}/applications/system-hub.desktop
install -Dpm0644 cmd/system-hub/Icon.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/system-hub.png

%files
%{_bindir}/system-hub
%{_datadir}/applications/system-hub.desktop
%{_datadir}/icons/hicolor/256x256/apps/system-hub.png

%changelog
* Tue May 19 2026 System Hub Team <dev@system-hub.example.com> - 0.1.0-1
- Initial Fedora RPM
