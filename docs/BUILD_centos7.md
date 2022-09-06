# Install requirements
```
yum install -y git make cmake3 g++ 
yum groupinstall "Development Tools"
```

## GCC 7 is needed to compile gRPC and pytorch
```
yum install centos-release-scl devtoolset-7-gcc*
```

## Enable GCC 7
```
scl enable devtoolset-7 bash
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
git clone git@github.com:tilsor/ModSecIntl_mod_wace.git ~/mod_wace
git clone git@github.com:tilsor/ModSecIntl_wace_core ~/wacecore
git clone git@github.com:tilsor/ModSecIntl_wace_plugins ~/plugins
git clone git@github.com:tilsor/ModSecIntl_roberta_model.git ~/roberta
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
systemctl restart httpd
```

## wacecore

1. Install go 1.16.15 and dependencies:

```
cd
wget https://storage.googleapis.com/golang/getgo/installer_linux
chmod u+x installer_linux
./installer_linux --version 1.16.15
source ~/.bash_profile
export PATH=$PATH:$GOPATH/bin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

2. Build wacecore:
```
cd ~/wacecore
make
go build
```

3. Build various plugins:
```
cd ~/plugins
make
cp model/*so ~/wacecore/plugins/model/
cp decision/*so ~/wacecore/plugins/decision/
```

4. Build Roberta plugin:
```
cd ~/roberta
make
cp roberta.so ~/wacecore/plugins/model
```

## Roberta model

Roberta needs at least Python 3.7:
```
cd
wget https://www.python.org/ftp/python/3.7.11/Python-3.7.11.tgz
tar xf Python-3.7.11.tgz
cd Python-3.7.11
./configure
make
make altinstall
```

It might be necessary to change python binary to `python3.7` in
`roberta/py/Makefile`.

```
cd ~/roberta/py
make
python3.7 -m pip install -r requirements.txt
```

# Run everything:
```
cd ~/wacecore
wace ./waceconfig.yaml &
cd ~/roberta/py
python3.7 server.py &
```

By default, wacecore listens on port 50051, but this can be changed in
configuration file (must change `WaceServerUrl` in
`/etc/httpd/conf/httpd.conf` too).

The Roberta model listens on port 9999. Can be changed in
`roberta/py/server.py`. Must also be changed in wacecore configuration
file, in the roberta plugin url field.
