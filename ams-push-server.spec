#debuginfo not supported with Go
%global debug_package %{nil}

Name: ams-push-server
Summary: ARGO Ams Push Server.
Version: 1.2.0
Release: 1%{?dist}
License: ASL 2.0
Buildroot: %{_tmppath}/%{name}-buildroot
Group: Unspecified
Source0: %{name}-%{version}.tar.gz
BuildRequires: golang
BuildRequires: git
Requires(pre): /usr/sbin/useradd, /usr/bin/getent
ExcludeArch: i386

%description
Installs the push server component

%pre
/usr/bin/getent group ams-push-server || /usr/sbin/groupadd -r ams-push-server
/usr/bin/getent passwd ams-push-server || /usr/sbin/useradd -r -s /sbin/nologin -d /var/www/ams-push-server -g ams-push-server ams-push-server

%prep
%setup

%build
export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin

cd src/github.com/ARGOeu/ams-push-server/
go install

%install
%{__rm} -rf %{buildroot}
install --directory %{buildroot}/var/www/ams-push-server
install --mode 755 bin/ams-push-server %{buildroot}/var/www/ams-push-server/ams-push-server

install --directory %{buildroot}/etc/ams-push-server
install --directory %{buildroot}/etc/ams-push-server/conf.d
install --mode 644 src/github.com/ARGOeu/ams-push-server/conf/ams-push-server-config.template %{buildroot}/etc/ams-push-server/conf.d/ams-push-server-config.json

install --directory %{buildroot}/usr/lib/systemd/system
install --mode 644 src/github.com/ARGOeu/ams-push-server/ams-push-server.service %{buildroot}/usr/lib/systemd/system/

%clean
%{__rm} -rf %{buildroot}
export GOPATH=$PWD
cd src/github.com/ARGOeu/ams-push-server/
go clean

%files
%defattr(0644,ams-push-server,ams-push-server)
%attr(0750,ams-push-server,ams-push-server) /var/www/ams-push-server
%attr(0755,ams-push-server,ams-push-server) /var/www/ams-push-server/ams-push-server
%config(noreplace) %attr(0644,ams-push-server,ams-push-server) /etc/ams-push-server/conf.d/ams-push-server-config.json
%attr(0644,root,root) /usr/lib/systemd/system/ams-push-server.service

%changelog
* Fri Jul 29 2022 Agelos Tsalapatis  <agelos.tsal@gmail.com> 1.2.0-1%{?dist}
- Release of ams-push-server 1.2.0
* Tue Oct 5 2021 Agelos Tsalapatis  <agelos.tsal@gmail.com> 1.0.1-1%{?dist}
- Release of ams-push-server 1.0.1
* Wed May 27 2020 Agelos Tsalapatis  <agelos.tsal@gmail.com> 1.0.0-1%{?dist}
- Release of ams-push-server 1.0.0
