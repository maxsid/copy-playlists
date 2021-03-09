package server

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/maxsid/playlists-copy/youtube"
	"golang.org/x/oauth2"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

// --- Errors Mock --- //

type errorsMock struct {
	errors []error
}

func (em *errorsMock) SetNextError(err ...error) {
	if em.errors == nil {
		em.errors = make([]error, 0)
	}
	em.errors = append(em.errors, err...)
}

// nextError returns the first element of em.errors and delete it from slice.
// Works by FIFO principe.
func (em *errorsMock) nextError() (err error) {
	if em == nil || em.errors == nil || len(em.errors) == 0 {
		return nil
	}
	err = em.errors[0]
	em.errors = em.errors[1:]
	return
}

// --- sessionsGetter --- //

type sessionsGetterMockT struct {
	sess sessionManager
	errorsMock
}

func (s *sessionsGetterMockT) Get(_ *fiber.Ctx) (sessionManager, error) {
	if err := s.nextError(); err != nil {
		return nil, err
	}
	return s.sess, nil
}

// ---- session Mock ---- //

type sessionMockT struct {
	*errorsMock
	Id        string
	Data      map[string]interface{}
	SaveCount int
}

func (s *sessionMockT) Destroy() error {
	if err := s.nextError(); err != nil {
		return err
	}
	s.Data = map[string]interface{}{}
	return nil
}

func newSessionMock(records map[string]interface{}, err ...error) *sessionMockT {
	if records == nil {
		records = make(map[string]interface{})
	}
	return &sessionMockT{Data: records, errorsMock: &errorsMock{errors: err}, Id: "mock-session-id"}
}

func (s *sessionMockT) ID() string {
	return s.Id
}

func (s *sessionMockT) Get(key string) (v interface{}) {
	v, _ = s.Data[key]
	return
}

func (s *sessionMockT) Set(key string, value interface{}) {
	s.Data[key] = value
}

func (s *sessionMockT) Save() error {
	if err := s.nextError(); err != nil {
		return err
	}
	s.SaveCount++
	return nil
}

func (s *sessionMockT) Delete(key string) {
	delete(s.Data, key)
}

func (s *sessionMockT) IsRecordExist(key string) bool {
	_, ok := s.Data[key]
	return ok
}

// ---- YouTubeUserServiceManagerCreator Mock ---- //

type youTubeUserServiceCreatorMockT struct {
	service youtube.Service
}

func newYouTubeUserServiceCreatorMockT(services ...youtube.Service) *youTubeUserServiceCreatorMockT {
	if len(services) > 0 {
		return &youTubeUserServiceCreatorMockT{service: services[0]}
	}
	return &youTubeUserServiceCreatorMockT{service: &youTubeUserServiceMockT{}}
}

func (u *youTubeUserServiceCreatorMockT) NewUserService() youtube.Service {
	return u.service
}

// ---- YouTubeUserServiceManager Mock ---- //

type youTubeUserServiceMockT struct {
	*errorsMock
	token         *oauth2.Token
	oauthConfig   youtube.Config
	userChannel   *youtubeAPI.Channel
	userPlaylists []*youtubeAPI.Playlist

	channels  []*youtubeAPI.Channel
	playlists []*youtubeAPI.Playlist
}

func newYouTubeUserServiceMockWithChannels(ch []*youtubeAPI.Channel, err ...error) *youTubeUserServiceMockT {
	if ch == nil {
		ch = make([]*youtubeAPI.Channel, 0)
	}
	return &youTubeUserServiceMockT{channels: ch, errorsMock: &errorsMock{errors: err}}
}

func (c *youTubeUserServiceMockT) PlaylistsByIDs(_ context.Context, id ...string) ([]*youtubeAPI.Playlist, error) {
	if err := c.nextError(); err != nil {
		return nil, err
	}
	idsMap := make(map[string]struct{})
	for _, i := range id {
		idsMap[i] = struct{}{}
	}

	playlists := make([]*youtubeAPI.Playlist, 0)
	for _, p := range c.playlists {
		if len(idsMap) == 0 {
			break
		}
		if _, ok := idsMap[p.Id]; ok {
			playlists = append(playlists, p)
			delete(idsMap, p.Id)
		}
	}
	return playlists, nil
}

func (c *youTubeUserServiceMockT) PlaylistByID(_ context.Context, id string) (*youtubeAPI.Playlist, error) {
	if err := c.nextError(); err != nil {
		return nil, err
	}
	for _, p := range c.playlists {
		if p.Id == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w playlist %s in mock", ErrNotFound, id)
}

func (c *youTubeUserServiceMockT) PlaylistsOfChannel(_ context.Context, channelID string) ([]*youtubeAPI.Playlist, error) {
	if err := c.nextError(); err != nil {
		return nil, err
	}
	ps := make([]*youtubeAPI.Playlist, 0)
	for _, p := range c.playlists {
		if p.Snippet.ChannelId == channelID {
			ps = append(ps, p)
		}
	}
	return ps, nil
}

func (c *youTubeUserServiceMockT) PlaylistItemsOfSeveralPlaylists(_ context.Context, _ ...string) ([]*youtubeAPI.PlaylistItem, error) {
	if err := c.nextError(); err != nil {
		return nil, err
	}
	return []*youtubeAPI.PlaylistItem{}, nil
}

func (c *youTubeUserServiceMockT) InsertPlaylistItems(_ context.Context, _ string, _ ...*youtubeAPI.PlaylistItem) ([]*youtubeAPI.PlaylistItem, error) {
	panic("implement me")
}

func (c *youTubeUserServiceMockT) ConfigUserService(_ context.Context, conf youtube.Config, tok *oauth2.Token) error {
	if err := c.nextError(); err != nil {
		return err
	}
	c.token, c.oauthConfig = tok, conf
	return nil
}

func (c *youTubeUserServiceMockT) ChannelOfMine(_ context.Context) (*youtubeAPI.Channel, error) {
	if err := c.nextError(); err != nil {
		return nil, err
	}
	ch := c.channels[len(c.channels)-1]
	c.channels = c.channels[:len(c.channels)-1]
	return ch, nil
}

// --- formValueGetter Mock --- //

type formValueGetterMockT struct {
	data map[string]string
}

func newFormValueGetterMockT(data map[string]string) *formValueGetterMockT {
	mock := &formValueGetterMockT{data: data}
	if data == nil {
		mock.data = make(map[string]string)
	}
	return mock
}

func (f *formValueGetterMockT) FormValue(key string, defaultValue ...string) string {
	if v, ok := f.data[key]; ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// --- youtube.Config mock --- //

type configMockT struct {
	url    string
	token  *oauth2.Token
	source oauth2.TokenSource
	errorsMock
}

func (c *configMockT) TokenSource(_ context.Context, _ *oauth2.Token) oauth2.TokenSource {
	return c.source
}

func (c *configMockT) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	return fmt.Sprintf("%s?state=%s", c.url, state)
}

func (c *configMockT) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if err := c.nextError(); err != nil {
		return nil, err
	}
	return c.token, nil
}
