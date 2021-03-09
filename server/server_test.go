package server

import (
	"errors"
	"fmt"
	"github.com/go-test/deep"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	youtubeAPI "google.golang.org/api/youtube/v3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
)

type testCase struct {
	doBeforeRequest       func(tc *testCase)
	requestURL            string
	requestMethod         string
	requestPostFormValues map[string]string
	session               *sessionMockT // session before request
	oauthConfig           *configMockT
	serviceCreator        *youTubeUserServiceCreatorMockT
	progressMapValue      interface{} // value in progressMap by session.ID() key before request.
	wantStatus            int
	wantSession           map[string]interface{} // want session data after request
	wantProgressMapValue  interface{}            // value in progressMap by session.ID() key after request.
	matchBodyPatterns     []string
}

func checkTestCase(t *testing.T, tc testCase, app *fiber.App) {
	progressMap = sync.Map{}
	if tc.session == nil {
		tc.session = newSessionMock(nil)
	}
	sessionStore = &sessionsGetterMockT{sess: tc.session}
	oauthConfig = tc.oauthConfig
	userServicesCreator = tc.serviceCreator

	if tc.progressMapValue != nil && tc.session != nil {
		progressMap.Store(tc.session.ID(), tc.progressMapValue)
	}

	if tc.doBeforeRequest != nil {
		tc.doBeforeRequest(&tc)
	}
	if tc.requestMethod == "" {
		tc.requestMethod = http.MethodGet
	}
	var formReader io.Reader
	if tc.requestPostFormValues != nil {
		values := url.Values{}
		for k, v := range tc.requestPostFormValues {
			values.Set(k, v)
		}
		formReader = strings.NewReader(values.Encode())
	}
	req, err := http.NewRequest(tc.requestMethod, tc.requestURL, formReader)
	if err != nil {
		panic(fmt.Errorf("request creating error: %w", err))
	}
	if formReader != nil {
		req.Header.Add(fiber.HeaderContentType, "application/x-www-form-urlencoded")
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Error(err)
		return
	}
	// check status code
	if resp.StatusCode != tc.wantStatus {
		t.Errorf("got status code = %d, want %d", resp.StatusCode, tc.wantStatus)
		return
	}
	// check session
	if tc.wantSession != nil {
		if diff := deep.Equal(tc.session.Data, tc.wantSession); diff != nil {
			t.Error(diff)
			return
		}
	}
	// check progressMapValue
	if tc.wantProgressMapValue != nil && tc.session != nil {
		v, _ := progressMap.Load(tc.session.ID())
		if diff := deep.Equal(v, tc.wantProgressMapValue); diff != nil {
			t.Error(diff)
			return
		}
	}
	// check page body
	if len(tc.matchBodyPatterns) > 0 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		for _, pattern := range tc.matchBodyPatterns {
			c := regexp.MustCompile(pattern)
			if !c.Match(body) {
				t.Errorf("Pattern '%s' hasn't match", pattern)
			}
		}
	}
}

func Test_auth(t *testing.T) {
	defaultToken := &oauth2.Token{AccessToken: "test_token"}
	app := createApp()

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "No state",
			tc: testCase{
				requestURL: "/auth?state=123456&code=12345678",
				session:    newSessionMock(map[string]interface{}{}),
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Empty code",
			tc: testCase{
				requestURL: "/auth?state=123456",
				session:    newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "123456"}),
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Save session error",
			tc: testCase{
				requestURL: "/auth?state=123456&code=12345678",
				session:    newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "123456"}, errors.New("save error")),
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Exchange error",
			tc: testCase{
				requestURL: "/auth?state=123456&code=12345678",
				session:    newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "123456"}),
				wantStatus: fiber.StatusInternalServerError,
				doBeforeRequest: func(tc *testCase) {
					tc.oauthConfig.SetNextError(errors.New("exchange error"))
				},
			},
		},
		{
			name: "OK",
			tc: testCase{
				requestURL:  "/auth?state=123456&code=12345678",
				session:     newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "123456"}),
				wantStatus:  fiber.StatusFound,
				wantSession: map[string]interface{}{sessionKeyOfYouTubeToken: defaultToken},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tc.oauthConfig = &configMockT{token: defaultToken}
			checkTestCase(t, tt.tc, app)
		})
	}
}

func Test_index(t *testing.T) {
	app := createApp()

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Success indexRequireAuthentication rendering",
			tc: testCase{
				requestURL:        "/",
				session:           newSessionMock(nil),
				oauthConfig:       &configMockT{url: "https://auth.example.com/auth"},
				wantStatus:        fiber.StatusOK,
				matchBodyPatterns: []string{`<title>Authenticate Youtube<\/title>`},
			},
		},
		{
			name: "Error token getting",
			tc: testCase{
				requestURL: "/",
				session:    newSessionMock(map[string]interface{}{sessionKeyOfYouTubeToken: 123455}),
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Success indexAuthenticated rendering",
			tc: testCase{
				requestURL: "/",
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken:     &oauth2.Token{AccessToken: "123456"},
					sessionKeyOfUserChannelCache: &youtubeAPI.Channel{Id: "654321", Snippet: &youtubeAPI.ChannelSnippet{Title: "testChannel"}},
				}),
				serviceCreator:    newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{}),
				wantStatus:        fiber.StatusOK,
				matchBodyPatterns: []string{`<title>Copy playlists</title>`},
			},
		},
		{
			name: "Wrong progress value",
			tc: testCase{
				requestURL: "/",
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken:     &oauth2.Token{AccessToken: "123456"},
					sessionKeyOfUserChannelCache: &youtubeAPI.Channel{Id: "654321", Snippet: &youtubeAPI.ChannelSnippet{Title: "testChannel"}},
				}),
				serviceCreator:   newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{}),
				progressMapValue: 123456,
				wantStatus:       fiber.StatusInternalServerError,
			},
		},
		{
			name: "Success indexInProgress rendering",
			tc: testCase{
				requestURL: "/",
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken:     &oauth2.Token{AccessToken: "123456"},
					sessionKeyOfUserChannelCache: &youtubeAPI.Channel{Id: "654321", Snippet: &youtubeAPI.ChannelSnippet{Title: "testChannel"}},
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "0", Snippet: &youtubeAPI.PlaylistSnippet{Title: "3"}},
						{Id: "1", Snippet: &youtubeAPI.PlaylistSnippet{Title: "4"}},
						{Id: "2", Snippet: &youtubeAPI.PlaylistSnippet{Title: "5"}},
					},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{}),
				progressMapValue: &copyingProgress{
					Count:        0,
					End:          10,
					DestPlaylist: &youtubeAPI.Playlist{Id: "31", Snippet: &youtubeAPI.PlaylistSnippet{Title: "play"}}},
				wantStatus: fiber.StatusOK,
				matchBodyPatterns: []string{
					`<title>Copying progress</title>`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.tc, app)
		})
	}
}

func Test_addPlaylists(t *testing.T) {
	app := createApp()

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "No playlists",
			tc: testCase{
				requestURL:            "/add",
				requestMethod:         fiber.MethodPost,
				requestPostFormValues: map[string]string{"links": "https://google.com/playlist?list=FL1FLtni3VzpVW9vjrFLF7UA"},
				wantStatus:            fiber.StatusFound,
			},
		},
		{
			name: "No token",
			tc: testCase{
				requestURL:            "/add",
				requestMethod:         fiber.MethodPost,
				requestPostFormValues: map[string]string{"links": "https://www.youtube.com/playlist?list=PL000001"},
				wantStatus:            fiber.StatusInternalServerError,
			},
		},
		{
			name: "Playlists getting error",
			tc: testCase{
				requestURL:            "/add",
				requestMethod:         fiber.MethodPost,
				requestPostFormValues: map[string]string{"links": "https://www.youtube.com/playlist?list=PL000001"},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "testToken"},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(newYouTubeUserServiceMockWithChannels(nil, nil, io.ErrNoProgress)),
				wantStatus:     fiber.StatusInternalServerError,
			},
		},
		{
			name: "Set source playlists error",
			tc: testCase{
				requestURL:            "/add",
				requestMethod:         fiber.MethodPost,
				requestPostFormValues: map[string]string{"links": "https://www.youtube.com/playlist?list=PL000001"},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "testToken"},
				}, ErrInvalidValue),
				serviceCreator: newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{playlists: []*youtubeAPI.Playlist{
					{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
				}}),
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Success",
			tc: testCase{
				requestURL:    "/add",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{"links": "https://www.youtube.com/playlist?list=PL000001\n" +
					"https://www.youtube.com/playlist?list=PL000002\n" +
					"https://www.youtube.com/playlist?list=PL000003\n"},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "testToken"},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{playlists: []*youtubeAPI.Playlist{
					{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
					{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
					{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
				}}),
				wantStatus: fiber.StatusFound,
				wantSession: map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
					},
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "testToken"},
				},
			},
		},
		{
			name: "Success with exists source playlists",
			tc: testCase{
				requestURL:            "/add",
				requestMethod:         fiber.MethodPost,
				requestPostFormValues: map[string]string{"links": "https://www.youtube.com/playlist?list=PL000003\n"},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "testToken"},
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
					},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{playlists: []*youtubeAPI.Playlist{
					{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
				}}),
				wantStatus: fiber.StatusFound,
				wantSession: map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
					},
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "testToken"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.tc, app)
		})
	}
}

func Test_deletePlaylists(t *testing.T) {
	app := createApp()

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Success",
			tc: testCase{
				requestURL:    "/delete",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{
					joinPlaylistDeletionPrefix(1): "on",
					joinPlaylistDeletionPrefix(3): "on",
				},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
						{Id: "PL000004", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000004"}},
						{Id: "PL000005", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000005"}},
					},
				}),
				wantStatus: fiber.StatusFound,
				wantSession: map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
						{Id: "PL000005", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000005"}},
					},
				},
			},
		},
		{
			name: "Save source playlists error",
			tc: testCase{
				requestURL:    "/delete",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{
					joinPlaylistDeletionPrefix(1): "on",
					joinPlaylistDeletionPrefix(3): "on",
				},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}},
						{Id: "PL000004", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000004"}},
						{Id: "PL000005", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000005"}},
					},
				}, ErrInvalidValue),
				wantStatus: fiber.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.tc, app)
		})
	}
}

func Test_startCopy(t *testing.T) {
	app := createApp()

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Success",
			tc: testCase{
				requestURL:    "/copy",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{
					"destination-playlist": "dest-playlist-id",
				},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "access-token"},
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
					},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{playlists: []*youtubeAPI.Playlist{
					{Id: "dest-playlist-id", Snippet: &youtubeAPI.PlaylistSnippet{Title: "dest-playlist-id"}},
				}}),
				progressMapValue: &copyingProgress{
					DestPlaylist: &youtubeAPI.Playlist{Id: "dest-playlist-id", Snippet: &youtubeAPI.PlaylistSnippet{Title: "dest-playlist-id"}},
					End:          15,
				},
				wantStatus: fiber.StatusFound,
			},
		},
		{
			name: "Error of service getting",
			tc: testCase{
				requestURL:    "/copy",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{
					"destination-playlist": "dest-playlist-id",
				},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "access-token"},
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
					},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(newYouTubeUserServiceMockWithChannels(nil, ErrInvalidValue)),
				progressMapValue: &copyingProgress{
					DestPlaylist: &youtubeAPI.Playlist{Id: "dest-playlist-id", Snippet: &youtubeAPI.PlaylistSnippet{Title: "dest-playlist-id"}},
					End:          15,
				},
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Error of getting destination playlist",
			tc: testCase{
				requestURL:    "/copy",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{
					"destination-playlist": "dest-playlist-id",
				},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "access-token"},
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
					},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(newYouTubeUserServiceMockWithChannels(nil, nil, ErrInvalidValue)),
				progressMapValue: &copyingProgress{
					DestPlaylist: &youtubeAPI.Playlist{Id: "dest-playlist-id", Snippet: &youtubeAPI.PlaylistSnippet{Title: "dest-playlist-id"}},
					End:          15,
				},
				wantStatus: fiber.StatusInternalServerError,
			},
		},
		{
			name: "Error of setCopyingProgress",
			tc: testCase{
				requestURL:    "/copy",
				requestMethod: fiber.MethodPost,
				requestPostFormValues: map[string]string{
					"destination-playlist": "dest-playlist-id",
				},
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "access-token"},
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
						{Id: "PL000003", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000003"}, ContentDetails: &youtubeAPI.PlaylistContentDetails{ItemCount: 5}},
					},
				}),
				serviceCreator: newYouTubeUserServiceCreatorMockT(&youTubeUserServiceMockT{playlists: []*youtubeAPI.Playlist{
					{Id: "dest-playlist-id", Snippet: &youtubeAPI.PlaylistSnippet{Title: "dest-playlist-id"}},
				}}),
				progressMapValue: &copyingProgress{
					DestPlaylist: &youtubeAPI.Playlist{Id: "dest-playlist-id", Snippet: &youtubeAPI.PlaylistSnippet{Title: "dest-playlist-id"}},
					End:          15,
				},
				doBeforeRequest: func(tc *testCase) {
					tc.session.Id = ""
				},
				wantStatus: fiber.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.tc, app)
			pr, ok := progressMap.Load(tt.tc.session.ID())
			if ok {
				if cp := pr.(*copyingProgress); cp.Cancel != nil {
					cp.Cancel()
				}
			}
		})
	}
}

func Test_stopCopy(t *testing.T) {
	app := createApp()
	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Success",
			tc: testCase{
				requestURL:    "/stop",
				requestMethod: fiber.MethodGet,
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
					},
				}),
				progressMapValue: &copyingProgress{Count: 10, End: 10},
				wantStatus:       fiber.StatusFound,
				wantSession:      map[string]interface{}{},
			},
		},
		{
			name: "Error of getting progress",
			tc: testCase{
				requestURL:    "/stop",
				requestMethod: fiber.MethodGet,
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
					},
				}),
				progressMapValue: 12345,
				wantStatus:       fiber.StatusInternalServerError,
			},
		},
		{
			name: "Error of deleting source playlists from session",
			tc: testCase{
				requestURL:    "/stop",
				requestMethod: fiber.MethodGet,
				session: newSessionMock(map[string]interface{}{
					sessionKeyOfSourcePlaylists: []*youtubeAPI.Playlist{
						{Id: "PL000001", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000001"}},
						{Id: "PL000002", Snippet: &youtubeAPI.PlaylistSnippet{Title: "Title PL000002"}},
					},
				}, ErrInvalidValue),
				progressMapValue: &copyingProgress{Count: 10, End: 10},
				wantStatus:       fiber.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.tc, app)
		})
	}
}

func Test_destroySession(t *testing.T) {
	app := createApp()

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Success",
			tc: testCase{
				requestURL:       "/destroy",
				requestMethod:    fiber.MethodGet,
				session:          newSessionMock(map[string]interface{}{"test1": 1, "test2": 2}),
				progressMapValue: &copyingProgress{Count: 10, End: 10},
				wantStatus:       fiber.StatusFound,
				wantSession:      map[string]interface{}{},
			},
		},
		{
			name: "Error of deleting progress",
			tc: testCase{
				requestURL:       "/destroy",
				requestMethod:    fiber.MethodGet,
				session:          newSessionMock(map[string]interface{}{"test1": 1, "test2": 2}),
				progressMapValue: 12345,
				wantStatus:       fiber.StatusInternalServerError,
			},
		},
		{
			name: "Success",
			tc: testCase{
				requestURL:       "/destroy",
				requestMethod:    fiber.MethodGet,
				session:          newSessionMock(map[string]interface{}{"test1": 1, "test2": 2}, ErrInvalidValue),
				progressMapValue: &copyingProgress{Count: 10, End: 10},
				wantStatus:       fiber.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.tc, app)
		})
	}
}
