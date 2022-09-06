# Install WACE

## wacecore
1. *TODO*: Add tilsor copr repo:
```
yum install -y yum-utils 
yum-config-manager --add-repo https://copr.fedorainfracloud.org/coprs/tilsor/modsecintl/repo/epel-8/tilsor-modsecintl-epel-8.repo
yum-config-manager --enable tilsor-modsecintl-epel-8
```

2. Install wacecore and roberta plugin rpms:
```
	dnf install -y wace wace-plugin-roberta
```


## Roberta model

1. Clone the repo:
```
git clone git@github.com:tilsor/ModSecIntl_roberta_model.git ~/roberta
```

2. Install dependencies:
```
dnf -y install python38 python38-devel

# There seems to be a bug with pip versions:
dnf -y install python38-numpy python38-Cython
```

3. 
```
cd ~/roberta/py
make
python3 -m pip install -r requirements.txt
```

# Run everything:
```
systemctl enable wace
systemctl enable roberta
systemctl start wace
systemctl start roberta
```

By default, wacecore listens on port 50051, but this can be changed in the wace
configuration file (`/etc/wace/waceconfig.yaml`). It must also be
changed in `/etc/httpd/conf/httpd.conf` with the option
`WaceServerUrl`.

The Roberta model listens on port 9999. It can be changed in
`config.ini`. It must also be changed in the wace
configuration file (`/etc/wace/waceconfig.yaml`), in the roberta
plugin url field.
