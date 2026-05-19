%global debug_package %{nil}

Name:           system-hub
Version:        0.1.0
Release:        alt1
Summary:        System administration GUI for Linux

License:        MIT
URL:            https://github.com/Lucky2356/system-hub
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.26
BuildRequires:  gcc
BuildRequires:  desktop-file-utils

Requires:       systemd
Requires:       docker-io

%description
System Hub is a desktop application (Fyne GUI) for monitoring and managing
a Linux system. Features:
  - Dashboard with CPU, RAM, disk, network, temperature
  - systemd service management (start/stop/restart/enable/disable)
  - Docker container and image management
  - Process viewer and port listing
  - Log viewer (journalctl, service logs, container logs)
  - File browser for configs and logs
  - Diagnostic report generation

%prep
%autosetup

%build
go build -trimpath -ldflags="-s -w" -o system-hub ./cmd/system-hub

%install
install -Dpm0755 system-hub %{buildroot}%{_bindir}/system-hub
install -Dpm0644 packaging/system-hub.desktop %{buildroot}%{_datadir}/applications/system-hub.desktop
install -Dpm0644 cmd/system-hub/Icon.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/system-hub.png

%post
%update_desktop_database
%update_icon_cache

%postun
%update_desktop_database
%update_icon_cache

%files
%{_bindir}/system-hub
%{_datadir}/applications/system-hub.desktop
%{_datadir}/icons/hicolor/256x256/apps/system-hub.png

%changelog
* Tue May 19 2026 System Hub Team <dev@system-hub.example.com> 0.1.0-alt1
- Initial Alt Linux package
