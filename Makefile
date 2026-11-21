.DEFAULT_GOAL := help
.PHONY: help lint lint-changed manifest-diff

help: ## Shows this overview
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

lint: ## Render all charts and run kubeconform + kube-linter (make lint CHARTS=apps/servarr)
	ci/lint.sh $(CHARTS)

lint-changed: ## Lint only charts changed since origin/main (make lint-changed BASE=main)
	ci/lint.sh --changed $(or $(BASE),origin/main)

manifest-diff: ## Diff rendered charts changed since origin/main (make manifest-diff BASE=main)
	@out=$$(mktemp -d); trap 'rm -rf "$$out"' EXIT; \
		LINT_OUT="$$out" ci/lint.sh --render-only --changed "$(or $(BASE),origin/main)" && \
		LINT_OUT="$$out" ci/manifest-diff.sh "$(or $(BASE),origin/main)"
