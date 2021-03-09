package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	middlewareCompress "github.com/gofiber/fiber/v2/middleware/compress"
	middlewareLogger "github.com/gofiber/fiber/v2/middleware/logger"
	middlewareRecover "github.com/gofiber/fiber/v2/middleware/recover"
	middlewareSession "github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/maxsid/playlists-copy/youtube"
	youtubeAPI "google.golang.org/api/youtube/v3"
	"log"
	"strings"
	"time"
)

var (
	sessionConfig = middlewareSession.Config{
		Expiration:     time.Hour,
		CookieName:     "session_id",
		KeyGenerator:   utils.UUIDv4,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}

	oauthConfig         youtube.Config
	sessionStore        sessionsGetter
	userServicesCreator youtube.ServiceCreator
)

// Run runs a web server.
func Run(addr string, conf youtube.Config, ysCreator youtube.ServiceCreator) {
	if conf == nil || ysCreator == nil {
		panic("Got nil conf or YouTubeUserServiceManagerCreator!")
	}
	oauthConfig, userServicesCreator = conf, ysCreator
	app := createApp()
	if err := app.Listen(addr); err != nil {
		panic(err)
	}
}

func createApp() *fiber.App {
	sessionStore = &sessionGettingStore{store: middlewareSession.New(sessionConfig)}

	engine, err := templatesEngine()
	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{Views: engine})

	initMiddlewares(app)
	initHandlers(app)

	return app
}

func initMiddlewares(app *fiber.App) {
	app.Use(middlewareLogger.New())
	app.Use(middlewareRecover.New())
	app.Use(middlewareCompress.New())
}

func initHandlers(app *fiber.App) {
	app.Get("/", index)
	app.Get("/auth", auth)
	app.Get("/destroy", destroySession)
	app.Post("/add", addPlaylists)
	app.Post("/delete", deletePlaylists)
	app.Post("/copy", startCopy)
	app.Get("/stop", stopCopy)
	app.Get("/static/*", static) // handles static
}

// index handles and renders "/" path.
func index(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)
	_, err := getAuthUserToken(sess)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return indexRequireAuthentication(c, sess, oauthConfig)
		}
		return err
	}
	progress, err := getCopyingProgress(sess.ID())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return indexAuthenticated(c, sess)
		}
		return err
	}
	return indexInProgress(c, sess, progress)
}

// auth handles GET "/auth" path which receives user's Google OAuth state and token.
func auth(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)
	data := getOAuthStateAndToken(c)
	if ok := compareUserAuthStates(sess, data.State); !ok {
		return fmt.Errorf("%w of states, their are not equal", ErrInvalidValue)
	}
	if data.Code == "" {
		return fmt.Errorf("%w of code: it's empty", ErrInvalidValue)
	}
	tok, err := oauthConfig.Exchange(context.TODO(), data.Code)
	if err != nil {
		return err
	}
	_ = setUserAuthState(sess, "", false) // save won't happen, so error won't happen
	if err = setAuthUserToken(sess, tok); err != nil {
		return err
	}

	return c.Redirect("/")
}

// addPlaylists handles "/add" path. Receives links from a textarea object and adds their into user's session record.
func addPlaylists(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)

	links := strings.Split(c.FormValue("links", ""), "\n")
	playlistsIds, _ := playlistsIDsFromLinks(links)
	if len(playlistsIds) == 0 {
		return c.Redirect("/")
	}
	serv, err := userService(context.TODO(), userServicesCreator, sess, oauthConfig)
	if err != nil {
		return err
	}

	newPlaylists, err := serv.PlaylistsByIDs(context.TODO(), playlistsIds...)
	if err != nil {
		return err
	}

	playlists := getSourcePlaylists(sess)
	if playlists == nil {
		playlists = newPlaylists
	} else {
		playlists = append(playlists, newPlaylists...)
	}

	if err = setSourcePlaylists(sess, playlists); err != nil {
		return err
	}
	return c.Redirect("/")
}

// deletePlaylists handles "/delete" path. Deletes concrete playlists from user's session record.
func deletePlaylists(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)
	playlists := getSourcePlaylists(sess)

	for pi, i := 0, 0; i < len(playlists); pi, i = pi+1, i+1 {
		got := c.FormValue(joinPlaylistDeletionPrefix(pi))
		if got == "on" {
			playlists = append(playlists[:i], playlists[i+1:]...)
			i--
		}
	}
	if err := setSourcePlaylists(sess, playlists); err != nil {
		panic(err)
	}
	return c.Redirect("/")
}

// startCopy handles "/copy" path. Starts copying of the selected playlists into user's playlist.
// Copying executes in separated goroutine copyPlaylists.
func startCopy(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)
	ctx, cancel := context.WithTimeout(c.Context(), time.Hour)
	playlists := getSourcePlaylists(sess)
	serv, err := userService(ctx, userServicesCreator, sess, oauthConfig)
	if err != nil {
		cancel()
		return err
	}
	destUserPlaylist, err := serv.PlaylistByID(ctx, c.FormValue("destination-playlist", ""))
	if err != nil {
		cancel()
		return err
	}
	progress := &copyingProgress{DestPlaylist: destUserPlaylist, End: countItemsOfPlaylists(playlists), Cancel: cancel}
	if err = setCopyingProgress(sess.ID(), progress); err != nil {
		cancel()
		return err
	}
	go copyPlaylists(ctx, cancel, sess.ID(), serv, playlists)
	return c.Redirect("/")
}

// stopCopy handles "/stop" path. Stops playlist copying and removes progress information.
func stopCopy(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)
	if err := deleteCopyingProgress(sess.ID()); err != nil {
		return err
	}
	if err := setSourcePlaylists(sess, nil); err != nil {
		return err
	}
	return c.Redirect("/")
}

// destroySession handles "/destroy" path. Destroys user's session.
func destroySession(c *fiber.Ctx) error {
	sess := mustSession(c, sessionStore)
	if err := deleteCopyingProgress(sess.ID()); err != nil {
		return err
	}
	if err := sess.Destroy(); err != nil {
		panic(err)
	}
	return c.Redirect("/")
}

// indexRequireAuthentication renders the index page if a user is not authenticated.
func indexRequireAuthentication(c *fiber.Ctx, sess sessionRecordSetterDeleterSaverDestroyer, urlGenerator authCodeURLGenerator) error {
	if err := sess.Destroy(); err != nil {
		return err
	}
	state := generateState()
	if err := setUserAuthState(sess, state); err != nil {
		return err
	}
	link := generateAuthLink(urlGenerator, state)
	return renderRequireAuth(c, link)
}

// indexAuthenticated renders the index page if a user is authenticated.
func indexAuthenticated(c *fiber.Ctx, sess sessionRecordGetterSetterSaver) error {
	serv, err := userService(context.TODO(), userServicesCreator, sess, oauthConfig)
	if err != nil {
		return err
	}
	userChannel, err := getUserChannel(context.TODO(), sess, serv)
	if err != nil {
		return err
	}
	userPlaylists, err := serv.PlaylistsOfChannel(context.TODO(), userChannel.Id)
	if err != nil {
		panic(err)
	}

	return renderIndex(c, renderIndexData{
		UserChannel:     userChannel,
		UserPlaylists:   userPlaylists,
		SourcePlaylists: getSourcePlaylists(sess),
	})
}

// indexInProgress renders index page if a user already ran playlists copying.
func indexInProgress(c *fiber.Ctx, sess sessionRecordGetterSetterSaver, progress *copyingProgress) error {
	serv, err := userService(context.TODO(), userServicesCreator, sess, oauthConfig)
	if err != nil {
		return err
	}
	userChannel, err := getUserChannel(context.TODO(), sess, serv)
	if err != nil {
		return err
	}
	sourcePlaylists := getSourcePlaylists(sess)
	if sourcePlaylists == nil {
		return fmt.Errorf("%w source playlists in the user session", ErrNotFound)
	}

	return renderProgress(c, renderProgressData{
		UserChannel:     userChannel,
		SourcePlaylists: sourcePlaylists,
		Progress:        progress,
	})
}

// copyPlaylists runs playlists copying. Information of Copying Progress  will be written into sync progressMap variable.
// When process is ended progress information won't be deleted.
func copyPlaylists(ctx context.Context, cancel context.CancelFunc, sessionID string, serv youtube.Service, playlists []*youtubeAPI.Playlist) {
	const updateStep = 10
	defer cancel()
	items, err := serv.PlaylistItemsOfSeveralPlaylists(ctx, playlistsIDsSlice(playlists)...)
	if err != nil {
		log.Println(err)
		return
	}
	progress, err := getCopyingProgress(sessionID)
	if err != nil {
		log.Println(err)
		return
	}

	if progress.End != len(items) {
		if err = setCopyingProgressEnd(sessionID, len(items)); err != nil {
			panic(err)
		}
	}

	for i := 0; i < len(items); i += updateStep {
		end := i + updateStep
		if end > len(items) {
			end = len(items)
		}
		if _, err = serv.InsertPlaylistItems(ctx, progress.DestPlaylist.Id, items[i:end]...); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Println(err)
			return
		}
		err = incrementCopyingProgress(sessionID, end-i)
		if err != nil {
			log.Println(err)
		}
	}
}
