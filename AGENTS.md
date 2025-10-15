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

## Commit Guidelines

### Commit Message Format

Use conventional commit message format for all commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Examples**:

- `feat(templates): add card format for styled content pages`
- `fix(generator): resolve asset versioning bug for CSS files`
- `refactor(build): extract template generation into separate command`

### Commit Organization

- **Group changes logically**: Each commit should contain related changes that serve a single purpose
- **Generated code separation**: Always commit generated code separately from source changes
- **Split generated code logically**:
  - Templ generated code (from `templates/*.templ.go` files)
  - Each type should be in its own commit
- **Asset processing**: Commit asset conversions (PNG to AVIF) separately when possible

### Commit Message Guidelines

- **Be detailed but concise**: Explain the intent and purpose, not every change
- **Focus on the "why"**: Describe the reason for the change, not just what was changed
- **Use imperative mood**: "Add feature" not "Added feature" or "Adding feature"

**Example workflow**:

```bash
# 1. Make application changes
git add cmd_generate.go
git commit -m "feat(generator): add support for markdown frontmatter parsing"

# 2. Commit Templ generated code
just gen-template
git add templates/*.templ.go
git commit -m "gen: update templ generated code for new layout components"

# 3. Commit asset conversions
just convert-images
git add pages/**/*.avif
git commit -m "assets: convert PNG images to AVIF format"
```

## Best Practices for AI Agents

### When Making Changes

1. **Template changes**: Always run `just gen-template` after modifying `.templ` files
2. **Build verification**: Run `just build` or `just build-dev` to verify changes work correctly
3. **Format code**: Use `just fmt` before committing
4. **Check dependencies**: Run `go mod tidy` when adding new Go dependencies
5. **Follow Git guidelines**: Use conventional commits and separate generated code
6. **Test builds**: Verify both production and development builds work when relevant

### Code Review Checklist

- [ ] All template files are properly formatted with `just fmt`
- [ ] Generated code is committed separately from source changes
- [ ] Frontmatter validation works correctly for new content types
- [ ] Asset versioning works in production builds
- [ ] Markdown content renders correctly with new templates
- [ ] Date formats match expected `"2006 January 02"` pattern
- [ ] Resume component IDs are correct and unique
- [ ] Build process completes without errors
- [ ] Development workflow (`just watch-build-dev`) functions properly

### Common Development Tasks

**Adding New Content Types**:
1. Define format in frontmatter processing logic
2. Create corresponding templ template if needed
3. Update documentation
4. Test with sample content

**Modifying Templates**:
1. Edit `.templ` files
2. Run `just gen-template`
3. Build and test changes
4. Commit generated code separately

**Asset Management**:
1. Add new assets to `assets/` directory
2. Update templates if referencing new assets
3. Test asset versioning in production builds
4. Convert images to AVIF when appropriate: `just convert-images`

## Code Style & Conventions

### Go Code Style

- **Formatting**: Use `just fmt` (runs `gofmt`, `goimports`, and `templ fmt`)
- **Naming**: CamelCase for exported identifiers, lowercase for unexported
- **Error handling**: Always check errors
- **Context**: Use context for cancellation and timeouts where appropriate

### Template Patterns

- **Formatting**: Use `just fmt` (includes `templ fmt`)
- **Templ files**: Located in `templates/`
- **Styling**: Custom CSS in `assets/`
- **JavaScript**: Minimal, mainly for syntax highlighting with Prism

### Content Patterns

- **Frontmatter**: YAML format with required fields based on content type
- **Date format**: Must be exactly `"2006 January 02"`
- **File organization**: Separate directories for different content types

## Troubleshooting

### Common Issues

1. **Template generation fails**:
   - Check for syntax errors in `.templ` files
   - Ensure templ CLI is installed: `go install github.com/a-h/templ/cmd/templ@latest`
2. **Build fails**:
   - Run `go mod tidy` to ensure dependencies are correct
   - Check that all `.templ` files have been generated: `just gen-template`
3. **Asset versioning issues**:
   - Verify assets are in correct directories
   - Check that CSS/JS files are properly referenced in templates
4. **Image conversion fails**:
   - Ensure ImageMagick is installed and `magick` command is available
   - Check that `fd` command is available for file finding
5. **Development watcher issues**:
   - Ensure `watchexec` is installed
   - Check file permissions and directory structure

## Additional Prerequisites

Based on the Justfile, the following additional tools are required:

- **[ImageMagick](https://imagemagick.org/)** - For PNG to AVIF conversion (`magick` command)
- **[fd](https://github.com/sharkdp/fd)** - Fast file finder (used in `convert-images` command)
- **[watchexec](https://github.com/watchexec/watchexec)** - File watcher for development workflows

## Deployment

The project includes deployment automation:

```bash
just deploy                    # Deploy to production server via rsync
```

Deployment assumes:
- SSH access to production server
- Proper SSH key configuration
- Correct server path in Justfile

## Project Structure

```
public-website/
├── main.go                   # CLI entry point using cobra
├── cmd_generate.go           # Core generation logic and content processing
├── templates/                # Templ template files and generated Go code
│   ├── *.templ              # Template source files
│   └── *.templ.go           # Generated Go code (don't edit)
├── pages/                   # Source Markdown content
│   ├── blog/                # Blog posts with frontmatter
│   ├── resume/              # Resume components
│   └── *.md                 # Standard pages
├── assets/                  # Static assets (CSS, JS)
├── files/                   # Static files copied to build
├── build/                   # Generated site output
├── Justfile                 # Build commands and development workflow
└── AGENTS.md                # This file - AI agent guidance
```

Remember to always start by ensuring dependencies are installed and run `just gen-template` after any template changes.
