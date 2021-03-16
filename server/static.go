package server

import (
	"embed"
	"github.com/gofiber/fiber/v2"
	"mime"
	"path/filepath"
)

//go:embed template/static
var staticDirectoryFS embed.FS

// static is a handler for static files. Uses '*' parameter for static file determining.
func static(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	path = filepath.Join("template", "static", path)
	file, err := staticDirectoryFS.Open(path)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	t := mime.TypeByExtension(filepath.Ext(path))
	if t == fiber.MIMEApplicationJavaScript { // add charset utf-8 for JS files, if it's not included
		t = fiber.MIMEApplicationJavaScriptCharsetUTF8
	}
	c.Set(fiber.HeaderContentType, t)
	c.Append(fiber.HeaderCacheControl, "public, max-age=31536000")
	return c.SendStream(file)
}
