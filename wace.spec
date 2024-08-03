Name:           wace
Version:        1.0
Release:        1%{?dist}
Summary:        A framework for adding machine learning capabilities to WAFs and OWASP CRS

License:        Apache v2.0
URL:            https://www.fing.edu.uy/inco/proyectos/wafmind
Source0:        %{name}-%{version}.tar.gz

BuildRequires: git
BuildRequires: golang
BuildRequires: systemd-rpm-macros
BuildRequires: protobuf-devel

%description 
A framework for adding machine learning capabilities to WAFs (such as
mod_security) and OWASP CRS. This package corresponds to the core
component that communicates mod_security with a ML model. The other
parts are a WAF module (such a mod_wace for Apache) and a machine
learning model plugin.

%global debug_package %{nil}

%prep
%autosetup
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

%build
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin

# This is only necessary as long as repos are private
go env -w GOPRIVATE=github.com/tilsor/ModSecIntl_logging
#cat << EOF > /tmp/github-credentials.sh
#!/bin/bash
#echo username=$GIT_USERNAME
#echo password=$GIT_PASSWORD
#EOF
#git config --global credential.helper "/bin/bash /tmp/github-credentials.sh"

make
go build -v -o %{name}

%install
install -Dpm 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -Dpm 0644 waceconfig.yaml %{buildroot}%{_sysconfdir}/%{name}/waceconfig.yaml
install -Dpm 644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service
install -Dpm 644 _plugins/decision/simple.so %{buildroot}%{_libdir}/%{name}/plugins/decision/simple.so

%check 
go test ./...

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%files
%dir %{_sysconfdir}/%{name}
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%config(noreplace) %{_sysconfdir}/%{name}/waceconfig.yaml
%{_libdir}/%{name}/plugins/decision/simple.so
# %license LICENSE

%changelog
* Tue Sep 6 2022 Juan Diego Campo <jdcampo@fing.edu.uy>
- Initial release 1.0-1
