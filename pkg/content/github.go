package content

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type GitHubContent struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
	Type    string `json:"type"`
	URL     string `json:"url"`
}

type GitHubFile struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (m *Manager) LoadPostsFromGitHub() error {
	repoURL := "https://api.github.com/repos/YOUR_GITHUB_USERNAME/blog-posts/contents"
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", repoURL + "/posts", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Blog-App")
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var files []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Printf("API Response: %+v\n", result)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	m.Posts = []Post{} // Reset posts list

	// Process each file
	for _, file := range files {
		if file.Type != "file" || filepath.Ext(file.Name) != ".md" {
			continue
		}

		// Fetch file content
		req, err := http.NewRequest("GET", file.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for file %s: %w", file.Name, err)
		}
		req.Header.Set("User-Agent", "Blog-App")
		
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch file %s: %w", file.Name, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GitHub API returned status %d for file %s", resp.StatusCode, file.Name)
		}

		var ghFile GitHubFile
		if err := json.NewDecoder(resp.Body).Decode(&ghFile); err != nil {
			return fmt.Errorf("failed to decode file content: %w", err)
		}

		// Decode base64 content
		content, err := base64.StdEncoding.DecodeString(ghFile.Content)
		if err != nil {
			return fmt.Errorf("failed to decode base64 content: %w", err)
		}

		// Process markdown content
		frontmatter, markdownContent, err := splitFrontmatter(content)
		if err != nil {
			return fmt.Errorf("frontmatter error: %w", err)
		}

		htmlContent, err := m.convertMarkdown(markdownContent)
		if err != nil {
			return fmt.Errorf("markdown conversion failed: %w", err)
		}

		post := Post{
			Slug:    strings.TrimSuffix(file.Name, ".md"),
			Title:   getStringWithDefault(frontmatter, "title", "Untitled Post"),
			Content: htmlContent,
			Excerpt: getStringWithDefault(frontmatter, "excerpt", ""),
		}

		// Handle date
		if rawDate, ok := frontmatter["date"]; ok {
			switch v := rawDate.(type) {
			case time.Time:
				post.Date = v
			case string:
				dateStr := strings.TrimSpace(v)
				if date, err := time.Parse("2006-01-02", dateStr); err == nil {
					post.Date = date
				} else {
					post.Date = time.Now()
				}
			default:
				post.Date = time.Now()
			}
		} else {
			post.Date = time.Now()
		}

		post.FormattedDate = post.Date.Format("January 2, 2006")
		m.Posts = append(m.Posts, post)
	}

	// Sort posts by date
	sort.Slice(m.Posts, func(i, j int) bool {
		return m.Posts[i].Date.After(m.Posts[j].Date)
	})

	return nil
} 