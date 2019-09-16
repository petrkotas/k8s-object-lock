CGO_ENABLED := 0 
GOOS := linux 

deploy: 
	kubectl apply -f ./manifests/lockvalidation-sa.yaml
	kubectl apply -f ./manifests/lockvalidation-cr.yaml
	kubectl apply -f ./manifests/lockvalidation-crb.yaml
	kubectl apply -f ./manifests/pod_lock-crd.yaml
	kubectl apply -f ./manifests/lockvalidation-dpl.yaml
	kubectl apply -f ./manifests/lockvalidation-svc.yaml
	kubectl apply -f ./manifests/lockvalidation-cfg.yaml
	kubectl label namespace default lockable=true

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

build-docker: 
	 docker build -t pkotas/lockvalidation . 

codegen:
	/usr/bin/env bash ./hack/update-codegen.sh

build: cmd/main.go 
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -a -installsuffix cgo -o lockvalidation -v $^

gen-cert:
	/usr/bin/env bash ./hack/gen_certs.sh 

clean: clean-manifest clean-bin

clean-manifest:
	rm ./manifests/lockvalidation-cfg.yaml

clean-bin:
	rm ./lockvalidation

.PHONY: deploy clean clean-manifest clean-bin gen-certs codegen undeploy deploy-local undeploy-local
