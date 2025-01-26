package main

import (
    "embed"
    "fmt"
    "html/template"
    "io/fs"
    "log"
    "time"

    "blog-site/pkg/config"
    "blog-site/pkg/content"
    "blog-site/pkg/handlers"

    "github.com/gin-gonic/gin"
)

//go:embed all:content/*
var contentFS embed.FS

//go:embed templates/*
var templateFS embed.FS

func main() {
    cfg, err := config.Load("config/config.toml")
    if err != nil {
        log.Fatal("Config error:", err)
    }

    // Create sub filesystems
    contentSubFS, err := fs.Sub(contentFS, "content")
    if err != nil {
        log.Fatal("Content FS error:", err)
    }

    templatesSubFS, err := fs.Sub(templateFS, "templates")
    if err != nil {
        log.Fatal("Template FS error:", err)
    }

    contentManager := content.New(contentSubFS, cfg)
    if err := contentManager.LoadPosts(); err != nil {
        log.Fatal("Posts load error:", err)
    }
    if err := contentManager.LoadAbout(); err != nil {
        log.Fatal("About page load error:", err)
    }

    h := handlers.New(contentManager, cfg)

    router := gin.Default()
    router.SetFuncMap(template.FuncMap{
        "safeHTML": func(html template.HTML) template.HTML { return html },
        "now":      func() time.Time { return time.Now() },
    })

    // Load templates
    templ := template.Must(template.New("").
        Funcs(router.FuncMap).
        ParseFS(templatesSubFS,
            "*.html",
            "partials/*.html",
        ))
    router.SetHTMLTemplate(templ)

    // Routes
    router.GET("/", h.Home)
    router.GET("/posts", h.Posts)
    router.GET("/posts/:slug", h.Post)
    router.GET("/about", h.About)

    // Server configuration
    log.Printf("Starting server on :%d", cfg.Server.Port)
    router.SetTrustedProxies(nil) // Remove proxy warning
    router.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
