# AGENTS.md

This file provides guidance to any agentic tool when working with code in this repository.

## Project Overview

This is a static site generator built in Go that powers Vincent Rischmann's personal website. The system processes Markdown files with YAML frontmatter to generate HTML pages, with special handling for blog posts and resume components.

## Development Commands

### Prerequisites
- Go 1.24+
- [just](https://github.com/casey/just) command runner
- [templ](https://github.com/a-h/templ) CLI tool for template generation

### Common Commands

```bash
# Setup and dependencies
go mod tidy                    # Install Go dependencies

# Template generation (required before building)
just gen-template              # Generate Go files from .templ templates

# Building
just build                     # Full production build with asset versioning
just build-dev                 # Development build (no asset versioning)
just clean                     # Clean build directory

# Development workflow
just watch-build-dev           # Watch files and rebuild automatically
just fmt                       # Format Go and templ code

# Docker development
just docker_dev                # Run with Docker and hot reload
```

## Architecture

### Core Components

**Static Site Generator (`cmd_generate.go:64-244`)**
- Main generation logic processes Markdown files in `pages/` directory
- Uses goldmark for Markdown parsing with YAML frontmatter support
- Implements asset versioning for cache busting (CSS, JS, AVIF files)
- Generates three types of content: standard pages, blog entries, and resume parts

**Content Processing Pipeline**
1. File collection: Walks `pages/` directory for `.md` files
2. Markdown parsing: Uses goldmark with metadata extension
3. Asset versioning: Renames assets with timestamp hash for cache busting
4. Template rendering: Uses templ for type-safe HTML generation

**Template System (`templates/`)**
- Uses templ (not standard Go templates) for type-safe HTML generation
- Template files (`.templ`) are compiled to Go code (`.go` files)
- Layout template provides consistent page structure with navigation
- Specialized templates for blog posts and resume pages

### Content Types

**Blog Posts (`pages/blog/`)**
- Require frontmatter: `title`, `description`, `date`, `format: blog_entry`
- Automatic table of contents generation using goldmark-toc
- Date format: `"2006 January 02"`
- Optional `require_prism: true` for syntax highlighting

**Resume Components (`pages/resume/`)**
- Modular markdown files with `format: resume_part`
- Use `id` field to specify component type:
  - `id: skills` - Skills section
  - `id: work_experience` - Work experience entries
  - `id: side_projects` - Side projects section

**Standard Pages**
- Regular content with `format: standard`
- Includes about page and code documentation

### Key File Locations

- `main.go` - CLI entry point using cobra
- `cmd_generate.go` - Core generation logic and content processing
- `templates/` - Templ template files and generated Go code
- `pages/` - Source Markdown content
- `assets/` - Static assets (CSS, JS)
- `build/` - Generated site output
- `Justfile` - Build commands and development workflow

## Development Workflow

1. **Template changes**: Run `just gen-template` to regenerate Go files from `.templ`
2. **Content changes**: Markdown files in `pages/` are processed automatically
3. **Asset changes**: CSS/JS files are versioned and copied to build directory
4. **Development server**: Use `just watch-build-dev` for automatic rebuilds

## Testing

No test files found in the codebase. This is a straightforward static site generator without complex business logic requiring extensive testing.

## Important Notes

- Always run `just gen-template` after modifying `.templ` files
- Asset versioning is disabled in development builds for faster iteration
- The system uses templ for templates, not standard Go html/template
- Date format in frontmatter must be exactly `"2006 January 02"`
- Resume components are automatically assembled based on `id` field values
