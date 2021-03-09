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
	c.Append(fiber.HeaderContentType, t)
	return c.SendStream(file)
}
