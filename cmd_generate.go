package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/a-h/templ"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/yuin/goldmark"
	goldmarkmeta "github.com/yuin/goldmark-meta"
	goldmarkast "github.com/yuin/goldmark/ast"
	goldmarkparser "github.com/yuin/goldmark/parser"
	goldmarkrenderer "github.com/yuin/goldmark/renderer"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	goldmarktoc "go.abhg.dev/goldmark/toc"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"go.rischmann.fr/website-generator/templates"
)

type generateCommandConfig struct {
	pagesDir  string
	assetsDir string
	buildDir  string

	logger *zap.Logger
}

func newGenerateCmd() *ffcli.Command {
	cfg := generateCommandConfig{
		logger: zap.L(),
	}

	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.StringVar(&cfg.pagesDir, "pages-directory", "./pages", "The directory where the markdown pages are stored")
	fs.StringVar(&cfg.assetsDir, "assets-directory", "assets", "The directory where the asset files are stored")
	fs.StringVar(&cfg.buildDir, "build-directory", "build", "The directory where the generated files will be stored")

	return &ffcli.Command{
		Name:       "generate",
		ShortUsage: "generate [flags]",
		ShortHelp:  "generate the website",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}
}

func (c *generateCommandConfig) Exec(ctx context.Context, args []string) error {
	generationDate := time.Now()

	// Copy the assets with the proper hash
	if err := c.copyAssets(ctx, generationDate); err != nil {
		return err
	}

	// Generating the website pages
	if err := c.generatePages(ctx, generationDate); err != nil {
		return err
	}

	return nil
}

func (c *generateCommandConfig) copyAssets(ctx context.Context, generationDate time.Time) error {
	c.logger.Info("copying assets")

	versionedExtensions := map[string]struct{}{
		".css": {},
		".js":  {},
	}

	err := filepath.WalkDir(c.assetsDir, func(inputPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		outputPath := inputPath

		// Rename versioned files
		ext := filepath.Ext(inputPath)
		if _, ok := versionedExtensions[ext]; ok {
			name := d.Name()
			name, _ = renameWithVersion(name, generationDate)
			outputPath = filepath.Join(filepath.Dir(inputPath), name)
		}

		// Copy file

		inputFile, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("unable to open file %q, err: %w", inputPath, err)
		}
		defer inputFile.Close()

		outputFile, err := createOutputFile(c.buildDir, outputPath)
		if err != nil {
			return fmt.Errorf("unable to create file %q, err: %w", inputPath, err)
		}
		defer outputFile.Close()

		if _, err := io.Copy(outputFile, inputFile); err != nil {
			return fmt.Errorf("unable to copy data, err: %w", err)
		}

		if err := outputFile.Sync(); err != nil {
			return fmt.Errorf("unable to sync output file, err: %w", err)
		}

		return nil
	})

	return err
}

func (c *generateCommandConfig) generatePages(ctx context.Context, generationDate time.Time) error {
	c.logger.Info("collecting pages")

	markdown := goldmark.New(
		goldmark.WithParserOptions(goldmarkparser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
		goldmark.WithExtensions(
			goldmarkmeta.Meta,
		),
	)

	// Collect pages
	allPages, err := collectPages(c.pagesDir, markdown.Parser())
	if err != nil {
		return fmt.Errorf("unable to collect pages, err: %w", err)
	}

	// Process pages
	for _, page := range allPages {
		if err := page.generate(c.logger, generationDate, markdown.Renderer(), c.buildDir); err != nil {
			return fmt.Errorf("unable to generate page, err: %w", err)
		}
	}

	// Generate the blog index page
	if err := generateBlogIndex(c.logger, generationDate, c.buildDir, allPages); err != nil {
		return fmt.Errorf("unable to generate blog index, err: %w", err)
	}

	// Generate the resume page
	if err := generateResume(c.logger, generationDate, markdown.Renderer(), c.buildDir, allPages); err != nil {
		return fmt.Errorf("unable to generate blog index, err: %w", err)
	}

	return nil
}

type markdownHTMLComponent struct {
	renderer goldmarkrenderer.Renderer
	source   []byte
	node     goldmarkast.Node
}

func (c markdownHTMLComponent) Render(ctx context.Context, w io.Writer) error {
	return c.renderer.Render(w, c.source, c.node)
}

var _ templ.Component = (*markdownHTMLComponent)(nil)

//

func createOutputFile(buildRootDir string, path string) (*os.File, error) {
	dir := filepath.Join(buildRootDir, filepath.Dir(path))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create directory tree, err: %w", err)
	}

	path = filepath.Join(buildRootDir, path)

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("unable to create output file %q, err: %w", path, err)
	}

	return f, nil
}

const (
	formatStandard   = "standard"
	formatBlogEntry  = "blog_entry"
	formatResumePart = "resume_part"
)

type pageMetadata struct {
	Title  string
	Date   time.Time
	Format string
	Extra  map[string]any
}

func parsePageMetadata(metadata map[string]any) (pageMetadata, error) {
	var res pageMetadata
	res.Extra = make(map[string]any)

	for k, v := range metadata {
		res.Extra[k] = v
	}

	if tmp, ok := res.Extra["title"]; ok {
		res.Title = tmp.(string)
	}

	if tmp, ok := res.Extra["date"]; ok {
		date, err := time.Parse("2006 January 02", tmp.(string))
		if err != nil {
			return pageMetadata{}, fmt.Errorf("invalid `date` value %q, should be a date in the `2006 Jan 02` format", tmp)
		}
		res.Date = date
	}

	if tmp, ok := res.Extra["format"]; ok {
		res.Format = tmp.(string)
	}

	return res, nil
}

// page represents a markdown page
type page struct {
	path       string // found while walking the pages root directory
	sourceData []byte // source bytes

	markdownDocument goldmarkast.Node // parsed from the source bytes
	metadata         pageMetadata     // found in the YAML header of the markdown page
}

func (p page) generate(logger *zap.Logger, generationDate time.Time, renderer goldmarkrenderer.Renderer, buildRootDir string) error {
	ctx := context.Background()

	assets := newAssets(generationDate)
	assets.add("style.css")

	if v, ok := p.metadata.Extra["require_prism"]; ok && v.(bool) {
		assets.add("prism.css")
		assets.add("prism.js")
	}

	//

	var page templ.Component
	switch p.metadata.Format {
	case formatStandard:
		content := markdownHTMLComponent{
			renderer: renderer,
			source:   p.sourceData,
			node:     p.markdownDocument,
		}

		page = templates.Page(p.metadata.Title, assets.underlying, content)

	case formatBlogEntry:
		// Generate the ToC
		toc, err := goldmarktoc.Inspect(p.markdownDocument, p.sourceData)
		if err != nil {
			return fmt.Errorf("unable to generate table of contents for page %s, err: %w", p.path, err)
		}

		tableOfContents := markdownHTMLComponent{
			renderer: renderer,
			source:   p.sourceData,
			node:     goldmarktoc.RenderList(toc),
		}

		content := markdownHTMLComponent{
			renderer: renderer,
			source:   p.sourceData,
			node:     p.markdownDocument,
		}

		blogContent := templates.BlogContent(p.metadata.Title, p.metadata.Date, tableOfContents, content)
		page = templates.Page(p.metadata.Title, assets.underlying, blogContent)

	default:
		logger.Debug("skipping page, unknown format", zap.String("path", p.path))
		return nil
	}

	// Rendering page

	f, err := createOutputFile(buildRootDir, p.path+".html")
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Info("generating file",
		zap.String("path", p.path),
		zap.String("output_path", f.Name()),
	)

	if err := page.Render(ctx, f); err != nil {
		return fmt.Errorf("unable to render page to file %q, err: %w", f.Name(), err)
	}

	return nil
}

type pages []page

func (p pages) getAll(format string) []page {
	res := make([]page, 0, len(p))
	for _, p := range p {
		if p.metadata.Format == format {
			res = append(res, p)
		}
	}
	return res
}

func collectPages(rootDir string, parser goldmarkparser.Parser) (res []page, err error) {
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		var page page

		// Parse and convert the page
		goldmarkContext := goldmarkparser.NewContext()
		{
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("unable to read file %q, err: %w", path, err)
			}
			document := parser.Parse(text.NewReader(data), goldmarkparser.WithContext(goldmarkContext))

			page.sourceData = data
			page.markdownDocument = document
		}

		// Parse the metadata from the markdown page
		{
			md, err := parsePageMetadata(goldmarkmeta.Get(goldmarkContext))
			if err != nil {
				return err
			}

			page.metadata = md
		}

		// Convert the path
		{
			relativePath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return fmt.Errorf("unable to get relative path of %s, err: %w", path, err)
			}

			ext := filepath.Ext(relativePath)
			path := relativePath[:len(relativePath)-len(ext)]

			page.path = path
		}

		res = append(res, page)

		return nil
	})

	return res, err
}

func generateBlogIndex(logger *zap.Logger, generationDate time.Time, buildRootDir string, pages pages) error {
	ctx := context.Background()

	assets := newAssets(generationDate)
	assets.add("style.css")

	// Generate the index page

	blogItemsPerYear := make(map[int][]templates.BlogItem)
	for _, page := range pages.getAll(formatBlogEntry) {
		year := page.metadata.Date.Year()

		items := blogItemsPerYear[year]
		items = append(items, templates.BlogItem{
			LinkURL:  page.path,
			LinkText: page.metadata.Title,
			Date:     page.metadata.Date,
		})

		blogItemsPerYear[year] = items
	}

	var blogItems []templates.BlogItems
	for year, items := range blogItemsPerYear {
		blogItems = append(blogItems, templates.BlogItems{
			Year:  year,
			Items: items,
		})
	}

	slices.SortFunc(blogItems, func(a, b templates.BlogItems) bool {
		return a.Year < b.Year
	})

	blogIndex := templates.BlogIndex(blogItems)
	page := templates.Page("Vincent Rischmann - Blog", assets.underlying, blogIndex)

	// Rendering page

	f, err := createOutputFile(buildRootDir, "blog.html")
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Info("generating blog index",
		zap.String("output_path", f.Name()),
	)

	if err := page.Render(ctx, f); err != nil {
		return fmt.Errorf("unable to render page to file %q, err: %w", f.Name(), err)
	}

	return nil
}

func generateResume(logger *zap.Logger, generationDate time.Time, render goldmarkrenderer.Renderer, buildRootDir string, pages pages) error {
	ctx := context.Background()

	assets := newAssets(generationDate)
	assets.add("style.css")

	// Build the resume compoentns

	var (
		skills       templ.Component
		experience   []templ.Component
		sideProjects templ.Component
	)

	resumeParts := pages.getAll(formatResumePart)
	for _, part := range resumeParts {
		id, ok := part.metadata.Extra["id"]
		if !ok {
			continue
		}

		switch id {
		case "skills":
			skills = markdownHTMLComponent{
				renderer: render,
				source:   part.sourceData,
				node:     part.markdownDocument,
			}

		case "work_experience":
			experience = append(experience, markdownHTMLComponent{
				renderer: render,
				source:   part.sourceData,
				node:     part.markdownDocument,
			})

		case "side_projects":
			sideProjects = markdownHTMLComponent{
				renderer: render,
				source:   part.sourceData,
				node:     part.markdownDocument,
			}
		}
	}

	resume := templates.Resume(skills, experience, sideProjects)
	page := templates.ResumePage("Vincent Rischmann - Resume", assets.underlying, resume)

	// Rendering page

	f, err := createOutputFile(buildRootDir, "resume.html")
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Info("generating resume",
		zap.String("output_path", f.Name()),
	)

	if err := page.Render(ctx, f); err != nil {
		return fmt.Errorf("unable to render page to file %q, err: %w", f.Name(), err)
	}

	return nil
}

type assets struct {
	generationDate time.Time
	underlying     templates.Assets
}

func newAssets(generationDate time.Time) *assets {
	res := new(assets)
	res.generationDate = generationDate
	return res
}

func (a *assets) add(name string) {
	newName, ext := renameWithVersion(name, a.generationDate)

	switch ext {
	case ".css":
		a.underlying.CSS = append(a.underlying.CSS, newName)
	case ".js":
		a.underlying.JS = append(a.underlying.JS, newName)
	default:
		panic(fmt.Errorf("invalid extension %q", ext))
	}
}

func renameWithVersion(name string, generationDate time.Time) (string, string) {
	ext := filepath.Ext(name)
	nameWithoutExt := name[:len(name)-len(ext)]

	newName := fmt.Sprintf("%s.%08x%s", nameWithoutExt, generationDate.Unix(), ext)

	return newName, ext
}
