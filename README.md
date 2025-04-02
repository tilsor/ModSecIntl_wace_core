# Web Attack Classification Engine (WACE)

The general objective of this project is to build machine
learning-assisted web application firewall mechanisms for the
identification, analysis and prevention of computer attacks on web
applications. The main idea is to combine the flexibility provided by
the classification procedures obtained from machine learning models
with the codified knowledge integrated in the specification of the
[OWASP Core Rule Set](https://coreruleset.org/) used by the [ModSecurity WAF](https://www.modsecurity.org/) to detect attacks, while
reducing false positives. The next figure shows a high-level
overview of the architecture:

![WACE architecture overview](https://github.com/tilsor/ModSecIntl_wace_server/blob/main/docs/images/architecture.jpg?raw=true "WACE architecture overview")

This repository contains WACE itself, the core component of the
solution. It connects ModSecurity to the machine learning models.

Please see the [Apache module
repo](https://github.com/tilsor/ModSecIntl_mod_wace) and the [machine
learning model
repo](https://github.com/tilsor/ModSecIntl_roberta_model) for the rest
of the components.

You can find more information about the project, including published
research articles, at the [WAF Mind
site](https://www.fing.edu.uy/inco/proyectos/wafmind)

## Installation
RPM packages for Red Hat Enterprise Linux 8 (or any compatible
distribution) are provided in the [releases
page](https://github.com/tilsor/ModSecIntl_wace_server/releases).

For compilation and manual installation instructions, please see the
[docs](https://github.com/tilsor/ModSecIntl_wace_server/tree/main/docs) directory.

Build RPM from source
```
cd ~/waceserver
rsync -av --progress . wace-{version} --exclude .git
tar -czvf {user}/rpmbuild/SOURCES/wace-{version}.tar.gz ./wace-{version}/
cd wace-{version}/
rpmbuild -ba wace.spec
```

## Licence
Copyright (c) 2022 Tilsor SA, Universidad de la República and
Universidad Católica del Uruguay. All rights reserved.

WACE and its components are distributed under Apache Software License
(ASL) version 2. Please see the enclosed LICENSE file for full
details.

