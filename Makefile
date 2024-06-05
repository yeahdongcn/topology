IMAGE ?= r0ckstar/slurmibtopology

.PHONY: all
all:
	go test ./...
	(cd pkg/slurm/topology/tree && go test -bench=.)
	go build .

.PHONY: image
image:
	docker buildx build --platform=linux/amd64,linux/arm64 --push -t $(IMAGE) ./docker