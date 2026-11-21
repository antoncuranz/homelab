.DEFAULT_GOAL := help
.PHONY: help lint

help: ## Shows this overview
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

lint: ## Render all charts and run kubeconform + kube-linter (make lint CHARTS=apps/servarr)
	ci/lint.sh $(CHARTS)
