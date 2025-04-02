# Install requirements
```
yum install -y git make cmake3 gcc g++ execstack
yum groupinstall "Development Tools"
```
 
## Install gRPC
```
git clone https://github.com/grpc/grpc.git $HOME/grpc
cd $HOME/grpc
git submodule update --init --recursive 
mkdir -p cmake/build && cd cmake/build
cmake3 ../..
make
make install
```

## Install Apache + ModSecurity + OWASP CRS
```
yum install libxml2-devel httpd httpd-devel mod_security
cd /etc/httpd/modsecurity.d/
wget https://github.com/coreruleset/coreruleset/archive/refs/tags/v3.3.2.tar.gz
tar xf v3.3.2.tar.gz
mv coreruleset-3.3.2 owasp-crs
cp owasp-crs/crs-setup.conf.example owasp-crs/crs-setup.conf
```

### Edit /etc/httpd/conf.d/mod_security.conf and add these lines:
```
IncludeOptional modsecurity.d/owasp-crs/*.conf
IncludeOptional modsecurity.d/owasp-crs/rules/*.conf
```


# Build and install WACE

## Clone WACE repos
```
git clone https://github.com/tilsor/ModSecIntl_mod_wace.git ~/mod_wace
git clone https://github.com/tilsor/ModSecIntl_wace_server.git ~/wacecore
```

## mod_wace 

```
cp ~/wacecore/wace.proto ~/mod_wace/wace.proto
cd ~/mod_wace
mkdir -p cmake/build
cd cmake/build 
cmake3 ../..
make

cp libgrpc_wace_client.so /usr/lib/
ldconfig 
apxs -Wl -Wc -cia -I/usr/include/libxml2 -I~/mod_wace -L~/mod_wace/cmake/build/ -lgrpc_wace_client ~/mod_wace/mod_wace.c 
cp ~/mod_wace/crs_rules/* /etc/httpd/modsecurity.d/owasp-crs/rules/
sed -i -e '$a\SecRuleRemoveById 949110' /etc/httpd/modsecurity.d/owasp-crs/rules/RESPONSE-999-EXCLUSION-RULES-AFTER-CRS.conf
sed -i -e '$a\WaceServerUrl localhost:50051' /etc/httpd/conf/httpd.conf
execstack -c /usr/lib64/httpd/modules/mod_wace.so
systemctl restart httpd
```

## wacecore

1. Install go 1.22.12 and dependencies:

```
cd
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

2. Build wacecore:
```
cd ~/wacecore
make
go build
```

# Run everything:
```
cd ~/wacecore
wace ./waceconfig.yaml &
```

By default, wacecore listens on port 50051, but this can be changed in
configuration file (must change `WaceServerUrl` in
`/etc/httpd/conf/httpd.conf` too).
