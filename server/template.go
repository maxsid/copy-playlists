package server

import (
	"embed"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"google.golang.org/api/youtube/v3"
	"io/fs"
	"net/http"
)

const (
	templateRequireAuth = "require_auth"
	templateIndex       = "index"
	templateProgress    = "progress"
)

//go:embed template/*.html
var templateFS embed.FS

// templatesEngine returns html.Engine for pages rendering.
func templatesEngine() (*html.Engine, error) {
	sfs, err := fs.Sub(templateFS, "template")
	if err != nil {
		return nil, err
	}
	engine := html.NewFileSystem(http.FS(sfs), ".html")
	return engine.AddFunc("GetThumbnailsUrl", getThumbnailsUrlOfPlaylistSnippet), nil
}

// renderRequireAuth renders page
func renderRequireAuth(c *fiber.Ctx, authLink string) error {
	return c.Render(templateRequireAuth, fiber.Map{
		"AuthLink": authLink,
	})
}

type renderIndexData struct {
	UserChannel     *youtube.Channel
	UserPlaylists   []*youtube.Playlist
	SourcePlaylists []*youtube.Playlist
}

// renderIndex renders index page
func renderIndex(c *fiber.Ctx, data renderIndexData) error {
	return c.Render(templateIndex, fiber.Map{
		"Channel":         data.UserChannel,
		"UserPlaylists":   data.UserPlaylists,
		"SourcePlaylists": data.SourcePlaylists,
		"ItemsCount":      countItemsOfPlaylists(data.SourcePlaylists),
	})
}

type renderProgressData struct {
	UserChannel     *youtube.Channel
	SourcePlaylists []*youtube.Playlist
	Progress        *copyingProgress
}

// renderProgress renders index page with copying progress.
func renderProgress(c *fiber.Ctx, data renderProgressData) error {
	return c.Render(templateProgress, fiber.Map{
		"Channel":         data.UserChannel,
		"DestPlaylist":    data.Progress.DestPlaylist,
		"SourcePlaylists": data.SourcePlaylists,
		"ItemsCount":      countItemsOfPlaylists(data.SourcePlaylists),
		"Progress":        data.Progress,
	})
}

// getThumbnailsUrlOfPlaylistSnippet returns URL of the medium size thumbnail of the playlist snippet.
// Returns "" if it's not specified.
func getThumbnailsUrlOfPlaylistSnippet(snippet *youtube.PlaylistSnippet) string {
	if snippet != nil && snippet.Thumbnails != nil && snippet.Thumbnails.Medium != nil {
		return snippet.Thumbnails.Medium.Url
	}
	return ""
}
