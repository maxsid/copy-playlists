package server

import (
	"context"
	"github.com/go-test/deep"
	"github.com/maxsid/playlists-copy/youtube"
	"golang.org/x/oauth2"
	youtubeAPI "google.golang.org/api/youtube/v3"
	"io"
	"net/http"
	"testing"
)

func Test_getUserChannel(t *testing.T) {
	type args struct {
		ctx  context.Context
		sess sessionRecordGetterSetterSaver
		serv youtube.ServiceChannelsGetter
	}
	tests := []struct {
		name    string
		args    args
		want    *youtubeAPI.Channel
		wantErr bool
	}{
		{
			name: "OK from cache",
			args: args{
				ctx:  context.TODO(),
				sess: newSessionMock(map[string]interface{}{sessionKeyOfUserChannelCache: &youtubeAPI.Channel{Id: "4321"}}),
				serv: newYouTubeUserServiceMockWithChannels(nil),
			},
			want: &youtubeAPI.Channel{Id: "4321"},
		},
		{
			name: "OK from serve",
			args: args{
				ctx:  context.TODO(),
				sess: newSessionMock(nil),
				serv: newYouTubeUserServiceMockWithChannels([]*youtubeAPI.Channel{{Id: "da4321"}}),
			},
			want: &youtubeAPI.Channel{Id: "da4321"},
		},
		{
			name: "Error invalid channel type",
			args: args{
				ctx:  context.TODO(),
				sess: newSessionMock(map[string]interface{}{sessionKeyOfUserChannelCache: 321}),
				serv: newYouTubeUserServiceMockWithChannels(nil),
			},
			wantErr: true,
		},
		{
			name: "Error serv.ChannelOfMine()",
			args: args{
				ctx:  context.TODO(),
				sess: newSessionMock(nil),
				serv: newYouTubeUserServiceMockWithChannels(nil, http.ErrBodyNotAllowed),
			},
			wantErr: true,
		},
		{
			name: "Error serv.Save()",
			args: args{
				ctx:  context.TODO(),
				sess: newSessionMock(nil, io.ErrShortBuffer),
				serv: newYouTubeUserServiceMockWithChannels([]*youtubeAPI.Channel{{Id: "da4321"}}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserChannel(tt.args.ctx, tt.args.sess, tt.args.serv)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_setUserAuthState(t *testing.T) {
	type args struct {
		sess     sessionRecordSetterDeleterSaver
		state    string
		makeSave []bool
	}
	tests := []struct {
		name        string
		args        args
		wantState   string
		wantDeleted bool
		wantErr     bool
	}{
		{
			name:      "OK Set",
			args:      args{sess: newSessionMock(nil), state: "abcdefg"},
			wantState: "abcdefg",
		},
		{
			name:        "OK Delete",
			args:        args{sess: newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "abcdefg"}), state: ""},
			wantDeleted: true,
		},
		{
			name:    "Save error",
			args:    args{sess: newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "abcdefg"}, io.ErrShortBuffer), state: "dsa"},
			wantErr: true,
		},
		{
			name: "Set without save",
			args: args{
				sess:     newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "abcdefg"}, io.ErrShortBuffer),
				state:    "dsa",
				makeSave: []bool{false},
			},
			wantState: "dsa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := tt.args.sess.(*sessionMockT)
			if err := setUserAuthState(sess, tt.args.state, tt.args.makeSave...); (err != nil) != tt.wantErr {
				t.Errorf("setUserAuthState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if sess.IsRecordExist(sessionKeyOfUserAuthState) == tt.wantDeleted {
				t.Errorf("wantDeleted=%t. tt.wantDeletedRecord must be deleted, but it's exist, or opposite", tt.wantDeleted)
				return
			}
			if tt.wantDeleted {
				return
			}
			if diff := deep.Equal(sess.Get(sessionKeyOfUserAuthState), tt.wantState); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_saveSession(t *testing.T) {
	type args struct {
		sess     sessionSaver
		makeSave []bool
	}
	tests := []struct {
		name          string
		args          args
		wantSaveCount int
		wantErr       bool
	}{
		{
			name:          "OK empty makeSave arg",
			args:          args{sess: newSessionMock(nil)},
			wantSaveCount: 1,
		},
		{
			name:          "OK one true makeSave arg",
			args:          args{sess: newSessionMock(nil), makeSave: []bool{true}},
			wantSaveCount: 1,
		},
		{
			name:          "OK without saving",
			args:          args{sess: newSessionMock(nil, io.ErrShortBuffer), makeSave: []bool{false, true}},
			wantSaveCount: 0,
		},
		{
			name:    "Error saving",
			args:    args{sess: newSessionMock(nil, io.ErrShortBuffer), makeSave: []bool{true, false}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := tt.args.sess.(*sessionMockT)
			if err := saveSession(sess, tt.args.makeSave...); (err != nil) != tt.wantErr {
				t.Errorf("saveSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && sess.SaveCount != tt.wantSaveCount {
				t.Errorf("wantSaveCount=%d, SaveCount=%d", tt.wantSaveCount, sess.SaveCount)
			}
		})
	}
}

func Test_compareUserAuthStates(t *testing.T) {
	type args struct {
		sess     sessionRecordGetter
		gotState string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Found, equal, True",
			args: args{sess: newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "12345"}), gotState: "12345"},
			want: true,
		},
		{
			name: "Found, not equal, False",
			args: args{sess: newSessionMock(map[string]interface{}{sessionKeyOfUserAuthState: "1234325"}), gotState: "12345"},
			want: false,
		},
		{
			name: "Not found, False",
			args: args{sess: newSessionMock(nil), gotState: "12345"},
			want: false,
		},
		{
			name: "Empty state, False",
			args: args{sess: newSessionMock(nil), gotState: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareUserAuthStates(tt.args.sess, tt.args.gotState); got != tt.want {
				t.Errorf("compareUserAuthStates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setAuthUserToken(t *testing.T) {
	type args struct {
		sess     sessionRecordSetterSaver
		token    *oauth2.Token
		makeSave []bool
	}
	tests := []struct {
		name          string
		args          args
		want          *oauth2.Token
		wantSaveCount int
		wantErr       bool
	}{
		{
			name:          "OK",
			args:          args{sess: newSessionMock(nil), token: &oauth2.Token{AccessToken: "123456789"}},
			wantSaveCount: 1,
			want:          &oauth2.Token{AccessToken: "123456789"},
		},
		{
			name: "OK without save",
			args: args{
				sess:     newSessionMock(nil, io.ErrShortBuffer),
				token:    &oauth2.Token{AccessToken: "123456789"},
				makeSave: []bool{false},
			},
			wantSaveCount: 0,
			want:          &oauth2.Token{AccessToken: "123456789"},
		},
		{
			name: "Error of saving",
			args: args{
				sess:     newSessionMock(nil, io.ErrShortBuffer),
				token:    &oauth2.Token{AccessToken: "123456789"},
				makeSave: []bool{true},
			},
			wantErr: true,
		},
		{
			name: "Error nil token",
			args: args{
				sess:     newSessionMock(nil),
				token:    nil,
				makeSave: []bool{true},
			},
			wantErr: true,
		},
		{
			name: "Error nil session",
			args: args{
				sess:     nil,
				token:    &oauth2.Token{AccessToken: "123456789"},
				makeSave: []bool{true},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setAuthUserToken(tt.args.sess, tt.args.token, tt.args.makeSave...); (err != nil) != tt.wantErr {
				t.Errorf("setAuthUserToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			sess := tt.args.sess.(*sessionMockT)
			if sess.SaveCount != tt.wantSaveCount {
				t.Errorf("wantSaveCount=%d, SaveCount=%d", tt.wantSaveCount, sess.SaveCount)
				return
			}
			if diff := deep.Equal(sess.Get(sessionKeyOfYouTubeToken), tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_getAuthUserToken(t *testing.T) {
	type args struct {
		sess sessionRecordGetter
	}
	tests := []struct {
		name    string
		args    args
		want    *oauth2.Token
		wantErr bool
	}{
		{
			name: "OK",
			args: args{sess: newSessionMock(map[string]interface{}{
				sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "12345678"},
			})},
			want: &oauth2.Token{AccessToken: "12345678"},
		},
		{
			name:    "Not found",
			args:    args{sess: newSessionMock(nil)},
			wantErr: true,
		},
		{
			name: "Wrong type",
			args: args{sess: newSessionMock(map[string]interface{}{
				sessionKeyOfYouTubeToken: 1234,
			})},
			wantErr: true,
		},
		{
			name: "Wrong value nil",
			args: args{sess: newSessionMock(map[string]interface{}{
				sessionKeyOfYouTubeToken: nil,
			})},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAuthUserToken(tt.args.sess)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAuthUserToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_userService(t *testing.T) {
	type args struct {
		ctx     context.Context
		creator *youTubeUserServiceCreatorMockT
		sess    sessionRecordGetter
		conf    *oauth2.Config
	}
	tests := []struct {
		name    string
		args    args
		want    *youTubeUserServiceMockT
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				ctx:     context.TODO(),
				creator: newYouTubeUserServiceCreatorMockT(),
				sess:    newSessionMock(map[string]interface{}{sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "12345"}}),
				conf:    &oauth2.Config{ClientID: "54321"},
			},
			want: &youTubeUserServiceMockT{
				token:       &oauth2.Token{AccessToken: "12345"},
				oauthConfig: &oauth2.Config{ClientID: "54321"},
			},
		},
		{
			name: "Error without token",
			args: args{
				ctx:     context.TODO(),
				creator: newYouTubeUserServiceCreatorMockT(),
				sess:    newSessionMock(nil),
				conf:    &oauth2.Config{ClientID: "54321"},
			},
			wantErr: true,
		},
		{
			name: "Error of user service configuration",
			args: args{
				ctx:     context.TODO(),
				creator: newYouTubeUserServiceCreatorMockT(newYouTubeUserServiceMockWithChannels(nil, io.ErrNoProgress)),
				sess:    newSessionMock(map[string]interface{}{sessionKeyOfYouTubeToken: &oauth2.Token{AccessToken: "12345"}}),
				conf:    &oauth2.Config{ClientID: "54321"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userService(tt.args.ctx, tt.args.creator, tt.args.sess, tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("userService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}
