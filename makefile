.PHONY: all
all: help

.PHONY: build-image
build-image:  ## build whereiszakir latest
	docker build . -t whereiszakir

.PHONY: upload-image
upload-image:  ## overwrite and upload whereiszakir latest image
	docker tag whereiszakir:latest dadrian/whereiszakir:latest
	docker push dadrian/whereiszakir:latest

.PHONY: build-linux
build-linux:  ## cross-compile a binary for linux
	GOOS=linux GOARCH=amd64 go build -o whereiszakir-linux

# via https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Show make help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
