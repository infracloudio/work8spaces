OPERATOR_SDK ?= operator-sdk

.PHONY: run deploy-crd

run: deploy-crd
	OPERATOR_NAME=work8spaces $(OPERATOR_SDK) run local --watch-namespace=""

deploy-crd:
	kubectl apply -f deploy/crds/work8space.infracloud.io_workspaces_crd.yaml
	kubectl apply -f deploy/crds/work8space.infracloud.io_workspaceusers_crd.yaml
