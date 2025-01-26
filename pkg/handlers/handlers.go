package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"blog-site/pkg/config"
	"blog-site/pkg/content"
)

type Handler struct {
	content *content.Manager
	cfg     *config.Config
}

func New(c *content.Manager, cfg *config.Config) *Handler {
	return &Handler{
		content: c,
		cfg:     cfg,
	}
}

func (h *Handler) Home(c *gin.Context) {
	// Get only the latest 4 posts for home page
	allPosts := h.content.Posts
	latestPosts := allPosts
	if len(allPosts) > 4 {
		latestPosts = allPosts[:4]
	}

	c.HTML(http.StatusOK, "base.html", gin.H{
		"Content": "home",
		"Site":    h.cfg.Site,
		"Posts":   latestPosts,
	})
}

func (h *Handler) Posts(c *gin.Context) {
	c.HTML(http.StatusOK, "base.html", gin.H{
		"Content": "posts",
		"Site":    h.cfg.Site,
		"Posts":   h.content.Posts,
	})
}

func (h *Handler) Post(c *gin.Context) {
	slug := c.Param("slug")
	for _, post := range h.content.Posts {
		if post.Slug == slug {
			c.HTML(http.StatusOK, "base.html", gin.H{
				"Content":  "post",
				"Post":     post,
				"Site":     h.cfg.Site,
			})
			return
		}
	}
	c.AbortWithStatus(http.StatusNotFound)
}

func (h *Handler) About(c *gin.Context) {
	c.HTML(http.StatusOK, "base.html", gin.H{
		"Content":  "about",
		"Site":     h.cfg.Site,
		"About":    h.content.About,
	})
}
