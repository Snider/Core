# core go mod graph

Print module dependency graph.

Wrapper around `go mod graph`. Outputs the module dependency graph in text form.

## Usage

```bash
core go mod graph
```

## What It Does

- Prints module dependencies as pairs
- Each line shows: `module@version dependency@version`
- Useful for understanding dependency relationships

## Examples

```bash
# Print dependency graph
core go mod graph

# Find who depends on a specific module
core go mod graph | grep "some/module"

# Visualise with graphviz
core go mod graph | dot -Tpng -o deps.png
```

## Output

```
github.com/host-uk/core github.com/stretchr/testify@v1.11.1
github.com/stretchr/testify@v1.11.1 github.com/davecgh/go-spew@v1.1.2
github.com/stretchr/testify@v1.11.1 github.com/pmezard/go-difflib@v1.0.1
...
```

## See Also

- [tidy](../tidy/) - Clean up go.mod
- [dev impact](../../../dev/impact/) - Show repo dependency impact
