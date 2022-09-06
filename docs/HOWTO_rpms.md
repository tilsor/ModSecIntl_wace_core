# Manual rpm generation with mock:

To generate rpm for a Linux distribution, first create a source rpm
from the spec file:

```
rpmbuild -bs wace.spec
```

Some dependency repos are private right now, so one must export the
environment variables `GIT_USERNAME` and `GIT_PASSWORD` with
appropriate values and use a mock configuration file similar to this:

```
include('templates/rocky-8.tpl')

config_opts['root'] = 'rocky-8-x86_64'
config_opts['target_arch'] = 'x86_64'
config_opts['legal_host_arches'] = ('x86_64',)
config_opts['environment']['GIT_USERNAME'] = os.environ['GIT_USERNAME']
config_opts['environment']['GIT_PASSWORD'] = os.environ['GIT_PASSWORD']
```

The spec file will use this variables to clone the private repos.

To build the binary rpm run:

```
mock --enable-network -r rocky8.cfg rebuild wace-0.1-1.fc34.src.rpm 
```

One can change rocky8 to the desired distribution, as long as it is
recent enough. 
