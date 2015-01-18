# ip2etcd
Posts Docker container IPs to etcd.  The current Linux x86 binary is available [here](https://github.com/colebrumley/ip2etcd/blob/master/bin/ip2etcd?raw=true).

```sh
Usage: ip2etcd [options]
Version: 0.1
  -c="": Target container (Shorthand)
  -container="": Target container
  -d="unix:///var/run/docker.sock": Docker socket or URL (Shorthand)
  -docker-endpoint="unix:///var/run/docker.sock": Docker socket or URL
  -e="http://127.0.0.1:4001": Comma separated list of etcd nodes (Shorthand)
  -etcd-nodes="http://127.0.0.1:4001": Comma separated list of etcd nodes
  -k="/net": Etcd base key (Shorthand)
  -key="/net": Etcd base key
  -q=false: Don't error when container IP is null
```
