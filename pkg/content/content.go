package content

import (
    "bytes"
    "fmt"
    "html/template"
    "io/fs"
    "time"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/parser"
    "gopkg.in/yaml.v3"
    "blog-site/pkg/config"
)

type Post struct {
    Slug          string
    Title         string
    Date          time.Time
    Excerpt       string
    Content       template.HTML
    FormattedDate string
}

type Manager struct {
    fs    fs.FS
    md    goldmark.Markdown
    cfg   *config.Config
    Posts []Post
    About template.HTML
}

func New(contentFS fs.FS, cfg *config.Config) *Manager {
    return &Manager{
        fs: contentFS,
        md: goldmark.New(
            goldmark.WithExtensions(extension.GFM),
            goldmark.WithParserOptions(
                parser.WithAutoHeadingID(),
            ),
        ),
        cfg: cfg,
    }
}

func splitFrontmatter(content []byte) (map[string]interface{}, []byte, error) {
    content = bytes.TrimSpace(content)

    // Handle both YAML and TOML style frontmatter delimiters
    var delimiter []byte
    if bytes.HasPrefix(content, []byte("---\n")) {
        delimiter = []byte("\n---\n")
    } else if bytes.HasPrefix(content, []byte("+++\n")) {
        delimiter = []byte("\n+++\n")
    } else {
        return nil, content, nil
    }

    split := bytes.SplitN(content, delimiter, 2)
    if len(split) < 2 {
        return nil, content, fmt.Errorf("invalid frontmatter format")
    }

    var frontmatter map[string]interface{}
    frontmatterContent := bytes.TrimPrefix(split[0], []byte("+++"))
    frontmatterContent = bytes.TrimPrefix(frontmatterContent, []byte("---"))

    // Debug: Print raw frontmatter content
    fmt.Printf("Raw frontmatter content: %s\n", string(frontmatterContent))

    if err := yaml.Unmarshal(frontmatterContent, &frontmatter); err != nil {
        return nil, content, fmt.Errorf("failed to parse frontmatter: %w", err)
    }

    // Debug: Print parsed frontmatter
    fmt.Printf("Parsed frontmatter: %+v\n", frontmatter)

    return frontmatter, split[1], nil
}

func (m *Manager) convertMarkdown(content []byte) (template.HTML, error) {
    var buf bytes.Buffer
    if err := m.md.Convert(content, &buf); err != nil {
        return "", fmt.Errorf("markdown conversion failed: %w", err)
    }
    return template.HTML(buf.String()), nil
}

func getStringWithDefault(fm map[string]interface{}, key string, def string) string {
    if val, ok := fm[key].(string); ok {
        return val
    }
    return def
}

func (m *Manager) LoadPosts() error {
    return m.LoadPostsFromGitHub()
}

func (m *Manager) LoadAbout() error {
    content, err := fs.ReadFile(m.fs, "about.md")
    if err != nil {
        return fmt.Errorf("reading about page: %w", err)
    }

    _, markdownContent, err := splitFrontmatter(content)
    if err != nil {
        return fmt.Errorf("about page frontmatter error: %w", err)
    }

    htmlContent, err := m.convertMarkdown(markdownContent)
    if err != nil {
        return fmt.Errorf("about page conversion failed: %w", err)
    }

    m.About = htmlContent
    return nil
}
