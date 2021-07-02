CGO_ENABLED := 0
GOOS := linux

deploy-cluster:
	kind create cluster

delete-cluster:
	kind delete cluster

cluster-load-image:
	kind load docker-image pkotas/lockvalidation:devel

deploy-base:
	kubectl apply -f ./manifests/lockvalidation-namespace.yaml
	kubectl apply -f ./manifests/lockvalidation-sa.yaml
	kubectl apply -f ./manifests/lockvalidation-cr.yaml
	kubectl apply -f ./manifests/lockvalidation-crb.yaml
	kubectl apply -f ./manifests/pod_lock-crd.yaml
	kubectl apply -f ./manifests/lockvalidation-svc.yaml
	kubectl apply -f ./manifests/lockvalidation-dpl.yaml
	kubectl label namespace default lockable=true

deploy-for-all:
	kubectl apply -f ./manifests/lockvalidation-cfg_all.yaml

deploy-for-dpl:
	kubectl apply -f ./manifests/lockvalidation-cfg_dpl.yaml

undeploy:
	kubectl delete -f ./manifests/lockvalidation-sa.yaml
	kubectl delete -f ./manifests/lockvalidation-cr.yaml
	kubectl delete -f ./manifests/lockvalidation-crb.yaml
	kubectl delete -f ./manifests/pod_lock-crd.yaml
	kubectl delete -f ./manifests/lockvalidation-dpl.yaml
	kubectl delete -f ./manifests/lockvalidation-svc.yaml
	kubectl delete -f ./manifests/lockvalidation-cfg.yaml
	kubectl label namespace default lockable-
	kubectl delete secret lockvalidation-crt

get-dependencies:
	go mod download

gen-cert:
	/usr/bin/env bash ./hack/gen_certs.sh --namespace kube-lock

gen-code:
	/usr/bin/env bash ./hack/update-codegen.sh

build-code: cmd/main.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -a -installsuffix cgo -o lockvalidation $^

build-docker:
	 docker build -t pkotas/lockvalidation:devel -f ./container/app.Dockerfile .

build: gen-code build-code build-docker

clean: clean-manifest clean-bin

clean-manifest:
	rm ./manifests/lockvalidation-cfg_all.yaml
	rm ./manifests/lockvalidation-cfg_dpl.yaml

clean-bin:
	rm ./lockvalidation

test-e2e:
	go test -tag e2e ./tests

test-unit:
	go test ./...

.PHONY: deploy-cluster deploy-to-cluster clean clean-manifest clean-bin gen-certs build codegen undeploy deploy-local undeploy-local get-dependencies
