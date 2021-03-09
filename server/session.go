package server

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/maxsid/playlists-copy/youtube"
	"golang.org/x/oauth2"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

const (
	sessionKeyOfYouTubeToken     = "youtube_token"
	sessionKeyOfUserAuthState    = "auth_state"
	sessionKeyOfUserChannelCache = "user_channel"
	sessionKeyOfSourcePlaylists  = "source_playlists"
)

// sessionStore is a wrap for session.Store object. Implements sessionsGetter.
type sessionGettingStore struct {
	store *session.Store
}

func (s *sessionGettingStore) Get(c *fiber.Ctx) (sessionManager, error) {
	if s.store == nil {
		return nil, fmt.Errorf("%w: sessionStore contains nil store", ErrInvalidValue)
	}
	return s.store.Get(c)
}

// saveSession saves session if the first parameter of makeSave is true or not specified session will be saved automatically.
// Otherwise all changes will be set and saving have to be done out of the function.
func saveSession(sess sessionSaver, makeSave ...bool) error {
	if len(makeSave) == 0 || len(makeSave) > 0 && makeSave[0] {
		return sess.Save()
	}
	return nil
}

// getAuthUserToken returns youtubeAPI.Service by token from session store.
func getAuthUserToken(sess sessionRecordGetter) (*oauth2.Token, error) {
	tokenInterface := sess.Get(sessionKeyOfYouTubeToken)
	if tokenInterface == nil {
		return nil, fmt.Errorf("token %w for this session", ErrNotFound)
	} else if token, ok := tokenInterface.(*oauth2.Token); ok {
		return token, nil
	}
	return nil, fmt.Errorf("%w YouTube user tokenInterface: saved an incorrect data (%T) in user session storage",
		ErrInvalidValue, tokenInterface)
}

// userService returns YouTubeUserService for current user by session.
func userService(ctx context.Context, creator youtube.ServiceCreator, sess sessionRecordGetter, conf youtube.Config) (youtube.Service, error) {
	tok, err := getAuthUserToken(sess)
	if err != nil {
		return nil, err
	}
	serv := creator.NewUserService()
	if err = serv.ConfigUserService(ctx, conf, tok); err != nil {
		return nil, err
	}
	return serv, nil
}

// setAuthUserToken saves YouTube token into user session.
// if makeSave is true or not specified session will be saved automatically.
// Otherwise a token will be set and saving have to be done out of the function.
func setAuthUserToken(sess sessionRecordSetterSaver, token *oauth2.Token, makeSave ...bool) error {
	if token == nil || sess == nil {
		return fmt.Errorf("%w of token or session. token=%v, session=%v", ErrInvalidValue, token, sess)
	}
	sess.Set(sessionKeyOfYouTubeToken, token)
	return saveSession(sess, makeSave...)
}

// compareUserAuthStates compares user authentication states from session and parameter. Returns true if they're equal.
func compareUserAuthStates(sess sessionRecordGetter, gotState string) bool {
	if gotState == "" {
		return false
	}
	storeState := sess.Get(sessionKeyOfUserAuthState)
	return storeState == gotState
}

// setUserAuthState stores user authentication state into session.
// For deleting record state have to be empty.
// if makeSave is true or not specified a session will be saved automatically.
// Otherwise a state will be set and saving have to be done out of the function.
func setUserAuthState(sess sessionRecordSetterDeleterSaver, state string, makeSave ...bool) error {
	if state == "" {
		sess.Delete(sessionKeyOfUserAuthState)
	} else {
		sess.Set(sessionKeyOfUserAuthState, state)
	}
	return saveSession(sess, makeSave...)
}

// getUserChannel gets user channel from cache or loads it from internet.
func getUserChannel(ctx context.Context, sess sessionRecordGetterSetterSaver, serv youtube.ServiceChannelsGetter) (ch *youtubeAPI.Channel, err error) {
	userChannelInterface := sess.Get(sessionKeyOfUserChannelCache)
	if userChannelInterface == nil {
		ch, err = serv.ChannelOfMine(ctx)
		if err != nil {
			return nil, err
		}
		sess.Set(sessionKeyOfUserChannelCache, ch)
		err = sess.Save()
		if err != nil {
			return nil, err
		}
		return
	}
	var ok bool
	ch, ok = userChannelInterface.(*youtubeAPI.Channel)
	if !ok {
		return nil, fmt.Errorf("%w of a channel from %s session record", ErrInvalidValue, sessionKeyOfUserChannelCache)
	}
	return
}

// getSourcePlaylists returns source playlists from a user session.
// If session doesn't have record function returns nil.
func getSourcePlaylists(sess sessionRecordGetter) []*youtubeAPI.Playlist {
	sourcePlaylists, _ := sess.Get(sessionKeyOfSourcePlaylists).([]*youtubeAPI.Playlist)
	return sourcePlaylists
}

// setSourcePlaylists sets source playlists in the session and saves data if makeSave parameter is set.
// If playlists parameter is nil, record will be deleted from the session.
func setSourcePlaylists(sess sessionRecordSetterDeleterSaver, playlists []*youtubeAPI.Playlist, makeSave ...bool) error {
	if playlists != nil {
		sess.Set(sessionKeyOfSourcePlaylists, playlists)
	} else {
		sess.Delete(sessionKeyOfSourcePlaylists)
	}
	return saveSession(sess, makeSave...)
}
