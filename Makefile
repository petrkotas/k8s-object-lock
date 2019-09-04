CGO_ENABLED := 0 
GOOS := linux 
GOPATH := /home/pkotas/Projects/Work/go/

deploy: 
	kubectl apply -f ./manifests/lockvalidation-sa.yaml
	kubectl apply -f ./manifests/lockvalidation-cr.yaml
	kubectl apply -f ./manifests/lockvalidation-crb.yaml
	kubectl apply -f ./manifests/pod_lock-crd.yaml
	kubectl apply -f ./manifests/lockvalidation-dpl.yaml
	kubectl apply -f ./manifests/lockvalidation-svc.yaml
	kubectl apply -f ./manifests/lockvalidation-cfg.yaml
	kubectl label namespace default lockable=true

build-docker: 
	 docker build -t pkotas/lockvalidation:v1 . 

codegen:
	/usr/bin/env bash ./hack/update-codegen.sh

build: cmd/main.go 
	GOPATH=$(GOPATH) CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -a -installsuffix cgo -o lockvalidation -v $^

gen-cert:
	/usr/bin/env bash ./hack/gen_certs.sh 

clean: clean-manifest clean-bin

clean-manifest:
	rm ./manifests/lockvalidation-svc.yaml

clean-bin:
	rm ./lockvalidation

.PHONY: deploy clean clean-manifest clean-bin gen-certs codegen
