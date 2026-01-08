# Docs Directory

This directory contains comprehensive documentation for the grule-plus project.

## Structure

```text
docs/
├── index.md          # Main documentation index
├── api.md            # API reference documentation
├── architecture.md   # System architecture and design
├── cache-types.md    # Cache algorithm explanations
├── configuration.md  # Configuration options and tuning
├── benchmarks.md     # Performance benchmarks and results
├── examples.md       # Usage examples and patterns
├── mkdocs.yml        # MkDocs configuration
└── diagrams/         # Architecture diagrams (future)
```

## Building Documentation

### Using MkDocs

1. Install MkDocs and plugins:

```bash
pip install mkdocs mkdocs-material mkdocstrings[python]
```

1. Serve documentation locally:

```bash
cd docs
mkdocs serve
```

1. Build static site:

```bash
cd docs
mkdocs build
```

### Using godoc

Generate Go documentation:

```bash
# Generate documentation for the project
godoc -http=:6060

# Or use go doc for specific packages
go doc ./engine
go doc ./internal/cache
```

## Documentation Guidelines

### API Documentation

- Use standard Go doc comments
- Include parameter descriptions
- Document return values and errors
- Provide usage examples

### Examples

- Include complete, runnable examples
- Cover common use cases
- Show error handling patterns
- Provide configuration examples

### Architecture Documentation

- Explain design decisions
- Include performance characteristics
- Document trade-offs
- Provide scaling guidelines

## Contributing to Documentation

1. Follow the existing structure and style
2. Use clear, concise language
3. Include code examples where helpful
4. Update the navigation in `mkdocs.yml` if adding new pages
5. Test documentation builds before submitting

## Tools Used

- **MkDocs**: Static site generator for documentation
- **Material Theme**: Modern, responsive theme
- **godoc**: Go standard documentation tool
- **Markdown**: Lightweight markup language
- **Mermaid** (future): For architecture diagrams
