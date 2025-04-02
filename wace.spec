Name:           wace
Version:        1.1
Release:        1%{?dist}
Summary:        A framework for adding machine learning capabilities to WAFs and OWASP CRS

License:        Apache v2.0
URL:            https://www.fing.edu.uy/inco/proyectos/wafmind
Source0:        %{name}-%{version}.tar.gz

BuildRequires: git
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
%autosetup -S git
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

%build
#export GOPATH=~/go
#export PATH=$PATH:$GOPATH/bin

# This is only necessary as long as repos are private
#go env -w GOPRIVATE=github.com/tilsor/ModSecIntl_logging
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
install -Dpm 644 _plugins/model/trivial.so %{buildroot}%{_libdir}/%{name}/plugins/model/trivial.so
install -Dpm 644 _plugins/model/trivial2.so %{buildroot}%{_libdir}/%{name}/plugins/model/trivial2.so
install -Dpm 644 _plugins/model/trivial_async.so %{buildroot}%{_libdir}/%{name}/plugins/model/trivial_async.so
install -Dpm 644 _plugins/decision/weighted_sum.so %{buildroot}%{_libdir}/%{name}/plugins/decision/weighted_sum.so

%check 
#go test ./...

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%files
%dir %{_sysconfdir}/%{name}
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%config(noreplace) %{_sysconfdir}/%{name}/waceconfig.yaml
%{_libdir}/%{name}/plugins/model/trivial.so
%{_libdir}/%{name}/plugins/model/trivial2.so
%{_libdir}/%{name}/plugins/model/trivial_async.so
%{_libdir}/%{name}/plugins/decision/weighted_sum.so
# %license LICENSE

%changelog
* Tue Sep 6 2022 Juan Diego Campo <jdcampo@fing.edu.uy>
- Initial release 1.0-1
