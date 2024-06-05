# topology

## Usage

### Host

```bash
make
./topology -p ./test/topology1.conf -a tux0 -a tux1 -a tux2 -c 3
./topology -p ./test/topology1.conf -a tux0 -a tux1 -a tux2 -a tux5 -a tux6 -r tux4 -c 3
./topology -p ./test/topology2.conf -a tux0 -a tux1 -a tux2 -c 3
./topology -p ./test/topology3.conf -a worker001 -a worker003 -a worker085 -a worker129 -a worker130 -a worker131 -c 3
```

### Docker

```bash
docker run --pull=always -it --privileged --net=host r0ckstar/slurmibtopology:latest
docker run --pull=always -it --privileged --net=host --entrypoint /bin/bash r0ckstar/slurmibtopology:latest
```

### Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```