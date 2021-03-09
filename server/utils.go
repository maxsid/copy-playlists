package server

import (
	"crypto/sha1"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/maxsid/playlists-copy/youtube/helper"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
	"math"
	"math/rand"
)

const (
	oauthGoogleGETVariableState = "state"
	oauthGoogleGETVariableCode  = "code"
)

// oauthRequestData contains request variables state and code after success Google authentication.
type oauthRequestData struct {
	State string
	Code  string
}

// getOAuthStateAndToken returns authentication result data from fiber.Ctx GET request.
func getOAuthStateAndToken(c formValueGetter) *oauthRequestData {
	return &oauthRequestData{
		Code:  c.FormValue(oauthGoogleGETVariableCode, ""),
		State: c.FormValue(oauthGoogleGETVariableState, ""),
	}
}

// generateAuthLink generates URL for Google authentication with state.
func generateAuthLink(generator authCodeURLGenerator, state string) string {
	return generator.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// generateState generates authenticate state by session user ID.
func generateState() string {
	bytesState := randBytes(sha1.BlockSize)
	return fmt.Sprintf("%x", sha1.Sum(bytesState))
}

// deletePlaylistsByIDs deletes playlists from slice by ID and returns cut slice.
func deletePlaylistsByIDs(playlist []*youtube.Playlist, ids ...string) []*youtube.Playlist {
	if len(playlist) == 0 || len(ids) == 0 {
		return playlist
	}
	idsMap := make(map[string]struct{})
	for _, id := range ids {
		idsMap[id] = struct{}{}
	}
	for i := 0; i < len(playlist) && len(ids) != 0; i++ {
		if _, ok := idsMap[playlist[i].Id]; ok {
			delete(idsMap, playlist[i].Id)
			playlist = append(playlist[:i], playlist[i+1:]...)
			i--
		}
	}
	return playlist
}

// joinPlaylistDeletionPrefix joins a prefix for a checkbox name on index page.
func joinPlaylistDeletionPrefix(playlistIndex int) string {
	return fmt.Sprintf("delete_%d", playlistIndex)
}

// playlistsIDsSlice returns IDs of all playlists objects.
func playlistsIDsSlice(playlists []*youtube.Playlist) []string {
	ids := make([]string, len(playlists))
	for i, p := range playlists {
		ids[i] = p.Id
	}
	return ids
}

// countItemsOfPlaylists adds items count of all playlists.
func countItemsOfPlaylists(playlists []*youtube.Playlist) int {
	sum := 0
	for _, p := range playlists {
		if p.ContentDetails != nil {
			sum += int(p.ContentDetails.ItemCount)
		}
	}
	return sum
}

// playlistsIDsFromLinks gets playlists IDs from their links.
// In the case of errors the function not interrupt and appends all errors into slice.
func playlistsIDsFromLinks(links []string) ([]string, []error) {
	playlistsIds := make([]string, 0)
	errs := make([]error, 0)
	for _, link := range links {
		id, err := helper.YoutubePlaylistIDFromURL(link)
		if err == nil {
			playlistsIds = append(playlistsIds, id)
		} else {
			errs = append(errs, err)
		}
	}
	return playlistsIds, errs
}

// mustSession is the same like session.Store.Get(), but executes a panic if got an error.
func mustSession(c *fiber.Ctx, store sessionsGetter) sessionManager {
	sess, err := store.Get(c)
	if err != nil {
		panic(err)
	}
	return sess
}

// randBytes generates random sliceSize bytes.
func randBytes(sliceSize int) []byte {
	bytes := make([]byte, sliceSize)
	for i := 0; i < sliceSize; i++ {
		bytes[i] = byte(rand.Intn(math.MaxInt8))
	}
	return bytes
}
