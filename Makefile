.PHONY: lint

lint: ## Render all charts and run kubeconform + kube-linter (make lint CHARTS=apps/servarr)
	ci/lint.sh $(CHARTS)
