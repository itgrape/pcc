package httpapi

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/pkg/types"
)

const webRoot = "/web"

func registerSPA(r *gin.Engine) error {
	if _, err := os.Stat(filepath.Join(webRoot, "index.html")); os.IsNotExist(err) {
		return nil
	}
	return registerSPAFrom(r, webRoot)
}

func registerSPAFrom(r *gin.Engine, webRoot string) error {
	root, err := filepath.Abs(webRoot)
	if err != nil {
		return fmt.Errorf("httpapi: resolve web root: %w", err)
	}
	indexPath := filepath.Join(root, "index.html")
	index, err := os.Stat(indexPath)
	if err != nil || !index.Mode().IsRegular() {
		return fmt.Errorf("httpapi: web root %q does not contain index.html", root)
	}

	static := http.FileServer(http.Dir(root))
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, types.Response{
				Code:    http.StatusNotFound,
				Message: "not found",
			})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}

		c.Header("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(path, "/assets/") || filepath.Ext(path) != "" {
			if strings.HasPrefix(path, "/assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Header("Cache-Control", "public, max-age=3600")
			}
			static.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Header("Cache-Control", "no-cache")
		c.File(indexPath)
	})
	return nil
}
