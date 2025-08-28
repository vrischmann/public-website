package main

import (
	"context"

	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	goldmarkmeta "github.com/yuin/goldmark-meta"
	goldmarkast "github.com/yuin/goldmark/ast"
	goldmarkparser "github.com/yuin/goldmark/parser"
	goldmarkrenderer "github.com/yuin/goldmark/renderer"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	goldmarktext "github.com/yuin/goldmark/text"
	goldmarkutil "github.com/yuin/goldmark/util"
	goldmarktoc "go.abhg.dev/goldmark/toc"
	"go.uber.org/multierr"

	"go.rischmann.fr/website-generator/templates"
)

type generateCommandConfig struct {
	pagesDir  string
	assetsDir string
	buildDir  string

	noAssetsVersioning bool

	logger *slog.Logger
}

func newGenerateCmd(logger *slog.Logger) *cobra.Command {
	cfg := &generateCommandConfig{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "generate the website",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cfg.Exec(context.Background(), args)
		},
	}

	cmd.Flags().StringVar(&cfg.pagesDir, "pages-directory", "./pages", "The directory where the markdown pages are stored")
	cmd.Flags().StringVar(&cfg.assetsDir, "assets-directory", "assets", "The directory where the asset files are stored")
	cmd.Flags().StringVar(&cfg.buildDir, "build-directory", "build", "The directory where the generated files will be stored")
	cmd.Flags().BoolVar(&cfg.noAssetsVersioning, "no-assets-versioning", false, "Disable assets versioning")

	return cmd
}

func (c *generateCommandConfig) Exec(ctx context.Context, args []string) error {
	var generationDate time.Time
	if !c.noAssetsVersioning {
		generationDate = time.Now()
	}

	// Copy all versioned files
	if err := c.copyVersionedFiles(ctx, generationDate); err != nil {
		return err
	}

	// Generating the website pages
	if err := c.generatePages(ctx, generationDate); err != nil {
		return err
	}

	return nil
}

func (c *generateCommandConfig) copyVersionedFiles(ctx context.Context, generationDate time.Time) error {
	c.logger.Info("copying files")

	versionedExtensions := map[string]struct{}{
		".css":  {},
		".js":   {},
		".avif": {},
	}

	doCopy := func(dir string, stripPrefix string) error {
		return filepath.WalkDir(dir, func(inputPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			outputPath := inputPath

			ext := filepath.Ext(inputPath)
			if _, ok := versionedExtensions[ext]; ok {
				// Rename and copy versioned files
				name := d.Name()

				if !generationDate.IsZero() {
					name, _ = renameWithVersion(name, generationDate)
				}

				dir := strings.TrimPrefix(filepath.Dir(inputPath), stripPrefix)
				outputPath = filepath.Join(dir, name)

			} else {
				// Not a file we want versioned, ignore
				return nil
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

			c.logger.Debug("copying file",
				slog.String("source", inputFile.Name()),
				slog.String("target", outputFile.Name()),
			)

			if _, err := io.Copy(outputFile, inputFile); err != nil {
				return fmt.Errorf("unable to copy data, err: %w", err)
			}

			if err := outputFile.Sync(); err != nil {
				return fmt.Errorf("unable to sync output file, err: %w", err)
			}

			return nil
		})
	}

	return multierr.Combine(
		doCopy(c.assetsDir, ""),
		doCopy(c.pagesDir, "pages/"),
	)
}

// imageVersioningTransformer is a goldmarkast.ASTTransformer that changes the images destination to include a hash of the generation date.
//
// This is needed for cache busting.
type imageVersioningTransformer struct {
	generationDate time.Time
}

func newImageVersioningTransformer(generationDate time.Time) *imageVersioningTransformer {
	return &imageVersioningTransformer{
		generationDate: generationDate,
	}
}

func (t *imageVersioningTransformer) Transform(node *goldmarkast.Document, reader goldmarktext.Reader, pc goldmarkparser.Context) {
	if t.generationDate.IsZero() {
		return
	}

	seen := make(map[*goldmarkast.Image]struct{})

	goldmarkast.Walk(node, func(n goldmarkast.Node, _ bool) (goldmarkast.WalkStatus, error) {
		img, ok := n.(*goldmarkast.Image)
		if !ok {
			return goldmarkast.WalkContinue, nil
		}

		if _, ok := seen[img]; ok {
			return goldmarkast.WalkContinue, nil
		}

		newFilename, _ := renameWithVersion(string(img.Destination), t.generationDate)
		img.Destination = []byte(newFilename)

		seen[img] = struct{}{}

		return goldmarkast.WalkContinue, nil
	})
}

var _ goldmarkparser.ASTTransformer = (*imageVersioningTransformer)(nil)

func (c *generateCommandConfig) generatePages(ctx context.Context, generationDate time.Time) error {
	c.logger.Info("collecting pages")

	markdown := goldmark.New(
		goldmark.WithParserOptions(
			goldmarkparser.WithAutoHeadingID(),
			goldmarkparser.WithASTTransformers(
				goldmarkutil.Prioritized(newImageVersioningTransformer(generationDate), 100),
			),
		),
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
	if c.node == nil {
		return nil
	}
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
	Title       string
	Description string
	Date        time.Time
	Format      string
	Extra       map[string]any
}

func parsePageMetadata(metadata map[string]any) (pageMetadata, error) {
	var res pageMetadata
	res.Extra = make(map[string]any)

	maps.Copy(res.Extra, metadata)

	if tmp, ok := res.Extra["title"]; ok {
		res.Title = tmp.(string)
	}

	if tmp, ok := res.Extra["description"]; ok {
		res.Description = tmp.(string)
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

func (p page) generate(logger *slog.Logger, generationDate time.Time, renderer goldmarkrenderer.Renderer, buildRootDir string) error {
	ctx := context.Background()

	assets := newAssets(generationDate)
	assets.add("style.css")
	assets.add("app.js")

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

		page = templates.Page(
			templates.HeaderParams{
				Title:       p.metadata.Title,
				Description: p.metadata.Description,
			},
			assets.underlying,
			content,
		)

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

		page = templates.Page(
			templates.HeaderParams{
				Title:       p.metadata.Title,
				Description: p.metadata.Description,
			},
			assets.underlying,
			blogContent,
		)

	default:
		logger.Debug("skipping page, unknown format", slog.String("path", p.path))
		return nil
	}

	// Rendering page

	f, err := createOutputFile(buildRootDir, p.path+".html")
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Info("generating file",
		slog.String("path", p.path),
		slog.String("output_path", f.Name()),
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
			document := parser.Parse(goldmarktext.NewReader(data),
				goldmarkparser.WithContext(goldmarkContext),
			)

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

func generateBlogIndex(logger *slog.Logger, generationDate time.Time, buildRootDir string, pages pages) error {
	ctx := context.Background()

	assets := newAssets(generationDate)
	assets.add("style.css")
	assets.add("app.js")

	// Generate the index page

	blogItemsPerYear := make(map[int][]templates.BlogItem)
	for _, page := range pages.getAll(formatBlogEntry) {
		year := page.metadata.Date.Year()

		linkURL := page.path
		if linkURL[0] != '/' {
			linkURL = "/" + linkURL
		}

		items := blogItemsPerYear[year]
		items = append(items, templates.BlogItem{
			LinkURL:  linkURL,
			LinkText: page.metadata.Title,
			Date:     page.metadata.Date,
		})

		blogItemsPerYear[year] = items
	}

	var blogItems []templates.BlogItems
	for year, items := range blogItemsPerYear {
		// Reverse sort, we want from most to least recent
		slices.SortFunc(items, func(a, b templates.BlogItem) int {
			return b.Date.Compare(a.Date)
		})

		blogItems = append(blogItems, templates.BlogItems{
			Year:  year,
			Items: items,
		})
	}

	slices.SortFunc(blogItems, func(a, b templates.BlogItems) int {
		return b.Year - a.Year
	})

	blogIndex := templates.BlogIndex(blogItems)
	page := templates.Page(
		templates.HeaderParams{
			Title:       "Vincent Rischmann - Blog",
			Description: "",
		},
		assets.underlying,
		blogIndex,
	)

	// Rendering page

	f, err := createOutputFile(buildRootDir, "blog.html")
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Info("generating blog index",
		slog.String("output_path", f.Name()),
	)

	if err := page.Render(ctx, f); err != nil {
		return fmt.Errorf("unable to render page to file %q, err: %w", f.Name(), err)
	}

	return nil
}

func generateResume(logger *slog.Logger, generationDate time.Time, render goldmarkrenderer.Renderer, buildRootDir string, pages pages) error {
	ctx := context.Background()

	assets := newAssets(generationDate)
	assets.add("style.css")
	assets.add("app.js")

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
	page := templates.ResumePage(
		templates.HeaderParams{
			Title:       "Vincent Rischmann - Resume",
			Description: "",
		},
		assets.underlying,
		resume,
	)

	// Rendering page

	f, err := createOutputFile(buildRootDir, "resume.html")
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Info("generating resume",
		slog.String("output_path", f.Name()),
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
	var newName, ext string
	if a.generationDate.IsZero() {
		ext = filepath.Ext(name)
		newName = name
	} else {
		newName, ext = renameWithVersion(name, a.generationDate)
	}

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
