# Personal Website Generator

A static site generator built in Go that powers my personal website at [https://rischmann.fr](https://rischmann.fr).

## Features

- **Static Site Generation**: Generates HTML from Markdown files with YAML frontmatter
- **Resume Builder**: Assembles resume from modular markdown components
- **Table of Contents**: Automatic TOC generation for blog posts
- **Responsive Design**: Clean, mobile-friendly design

## Tech Stack

- **Language**: Go 1.24.1
- **Templating**: [templ](https://github.com/a-h/templ) for HTML templates
- **Markdown**: [goldmark](https://github.com/yuin/goldmark) for Markdown processing
- **Styling**: Vanilla CSS with Prism.js for syntax highlighting
- **Web Server**: Caddy 2
- **Build Tool**: [just](https://github.com/casey/just)

## Project Structure

```
├── pages/                 # Content source files
│   ├── blog/             # Blog posts (markdown)
│   ├── code/             # Code documentation
│   ├── resume/           # Resume components
│   ├── about.md          # About page
│   └── code.md           # Code page
├── templates/            # HTML templates (templ)
├── assets/               # Static assets (CSS, JS)
├── files/                # Static files (PDFs, images)
├── build/                # Generated site (output)
├── cmd_generate.go       # Main generation logic
├── main.go              # CLI entry point
└── Dockerfile           # Container configuration
```

## Content Types

### Blog Posts
- Located in `pages/blog/`
- Require YAML frontmatter with `title`, `description`, `date`, and `format: blog_entry`
- Support automatic table of contents generation
- Images are automatically versioned for cache busting

### Resume Components
- Modular markdown files in `pages/resume/`
- Automatically assembled into a single resume page
- Components: skills, work experience, side projects

### Standard Pages
- Regular markdown pages with `format: standard`
- Includes about page and code documentation

## Development

### Prerequisites
- Go 1.24.1+
- [just](https://github.com/casey/just) command runner
- [templ](https://github.com/a-h/templ) CLI tool
- Optional: watchexec for file watching

### Quick Start

```bash
# Install dependencies
go mod tidy

# Generate templates
just gen-template

# Build the site
just build

# Development build (no asset versioning)
just build-dev

# Watch for changes and rebuild
just watch-build-dev
```

### Available Commands

```bash
# Build commands
just build          # Full production build
just build-dev      # Development build (no versioning)
just clean          # Clean build directory

# Development
just watch-build    # Watch and build production
just watch-build-dev # Watch and build development
just watch-convert-images # Watch PNG files and convert to AVIF

# Code formatting
just fmt            # Format Go and templ code

# Docker
just docker_dev     # Run in Docker with hot reload

# Deployment
just deploy         # Deploy to production server
```

### Content Management

#### Adding a Blog Post
1. Create a new `.md` file in `pages/blog/`
2. Add YAML frontmatter:
   ```yaml
   title: "Your Post Title"
   description: "Brief description"
   date: "2024 January 15"
   format: blog_entry
   require_prism: true  # Optional: for code highlighting
   ```
3. Write your content in Markdown

#### Adding Resume Content
1. Add/modify files in `pages/resume/`
2. Use the `id` field in YAML frontmatter to specify component type:
   - `id: skills` - Skills section
   - `id: work_experience` - Work experience entries
   - `id: side_projects` - Side projects section

## Deployment

### CI/CD Pipeline (Recommended)
The project includes a GitHub Actions workflow that automatically builds and deploys your website when you push to the main branch.

#### Setup GitHub Actions
1. Add these secrets to your GitHub repository:
   - `DEPLOY_HOST`: Your server hostname (e.g., `wevo.rischmann.fr`)
   - `DEPLOY_USER`: SSH username for deployment
   - `DEPLOY_KEY`: Private SSH key for authentication

2. Generate an SSH key pair for deployment:
   ```bash
   ssh-keygen -t rsa -b 4096 -C "github-actions@yourdomain.com" -f ~/.ssh/github_deploy
   ```

3. Add the public key to your server's `~/.ssh/authorized_keys`
4. Add the private key as `DEPLOY_KEY` secret in GitHub

#### Manual Deployment Script
```bash
# Build the site
just build

# Deploy to server (requires SSH access)
just deploy
```

The deployment script (`deploy.sh`) provides better error handling and verification compared to the basic rsync command.

### Docker
```bash
# Build and run with Docker Compose
docker compose up --build

# Development with hot reload
just docker_dev
```

## Configuration

### Environment Variables
- No environment variables required for basic usage
- Caddy configuration in `Caddyfile`
- Docker configuration in `compose.yaml`

### Customization
- Templates: Edit `.templ` files in `templates/`
- Styling: Modify `assets/style.css`
- Syntax highlighting: Update `assets/prism.css` and `assets/prism.js`

## License

This project is open source and available under the [MIT License](LICENSE).
