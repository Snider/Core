.PHONY: prod-docs

all:
	(cd cmd/app && wails3 build)

development-docs:
	@echo "Running development documentation Website..."
	@(cd pkg/core/docs && mkdocs serve -w src)

prod-docs:
	@echo "Generating documentation tp Repo Root..."
	@(cd pkg/core/docs && mkdocs build -d public && cp -r src public)
	@echo "Documentation generated at docs/index.html"