# VapusAI Documentation Service

A Docusaurus-based documentation site for VapusAI, providing comprehensive documentation for all platform features, APIs, and user guides.

## Development

### Prerequisites
- Node.js 18+ 
- npm

### Quick Start

```bash
# Install dependencies
make install

# Start development server
make dev
```

The documentation will be available at http://localhost:3000

### Building for Production

```bash
# Build the documentation
make build-docs

# Serve the built documentation
make serve-local
```

### Docker

```bash
# Build Docker image
make build

# Run in Docker
make run
```

The containerized documentation will be available at http://localhost:3000

## Project Structure

```
docs/
├── docs/           # Documentation pages
├── src/            # Custom React components and pages
├── static/         # Static assets
├── docusaurus.config.js  # Docusaurus configuration
└── package.json    # Dependencies and scripts
```

## Adding Documentation

1. Create markdown files in the `docs/` directory
2. Update `docusaurus.config.js` for navigation
3. Add static assets to `static/` directory
4. Create custom pages in `src/pages/`

## Integration with VapusAI

This documentation service is designed to run alongside other VapusAI services and can be accessed through the main application or directly via its port.
