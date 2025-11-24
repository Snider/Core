.PHONY: all dev prod-docs development-docs

all:
	(cd cmd/core-gui && task build)

.ONESHELL:
dev:
	(cd cmd/core-gui && task dev)

pre-commit:
	coderabbit review --prompt-only

development-docs:
	@echo "Running development documentation Website..."
	@(cd pkg/core/docs && mkdocs serve -w src)

prod-docs:
	@echo "Generating documentation tp Repo Root..."
	@(cd pkg/core/docs && mkdocs build -d public && cp -r src public)
	@echo "Documentation generated at docs/index.html"