# ip2etcd
Posts Docker container IPs to etcd.  The current Linux x64 binary is available [here](https://github.com/colebrumley/ip2etcd/blob/master/bin/ip2etcd?raw=true).  It works but I need to clean up the code some.  Right now it posts container IPs to `/key/[container name or short ID]/ip`.

```sh
---------------------
Usage: ip2etcd [options] containers
Version: 0.2
---------------------
  -a=false: Update [a]ll containers
  -d="unix:///var/run/docker.sock": [d]ocker endpoint
  -e="http://127.0.0.1:4001": Comma separated list of [e]tcd endpoints
  -k="/test": etcd base [k]ey
---------------------
```
