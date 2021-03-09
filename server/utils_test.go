package server

import (
	"github.com/go-test/deep"
	"google.golang.org/api/youtube/v3"
	"math/rand"
	"reflect"
	"testing"
)

func Test_deletePlaylistsByIDs(t *testing.T) {
	type args struct {
		playlist []*youtube.Playlist
		ids      []string
	}
	tests := []struct {
		name string
		args args
		want []*youtube.Playlist
	}{
		{
			name: "Empty playlists",
			args: args{playlist: []*youtube.Playlist{}, ids: []string{"1", "2", "3", "4", "5", "6", "7"}},
			want: []*youtube.Playlist{},
		},
		{
			name: "Empty ids",
			args: args{
				playlist: []*youtube.Playlist{{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}, {Id: "5"}},
				ids:      []string{},
			},
			want: []*youtube.Playlist{{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}, {Id: "5"}},
		},
		{
			name: "Success deleting",
			args: args{
				playlist: []*youtube.Playlist{{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}, {Id: "5"}},
				ids:      []string{"3", "5", "1"},
			},
			want: []*youtube.Playlist{{Id: "2"}, {Id: "4"}},
		},
		{
			name: "Not found",
			args: args{
				playlist: []*youtube.Playlist{{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}, {Id: "5"}},
				ids:      []string{"7", "8", "9", "10"},
			},
			want: []*youtube.Playlist{{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}, {Id: "5"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deletePlaylistsByIDs(tt.args.playlist, tt.args.ids...)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("deletePlaylistsByIDs() -> %v", diff)
			}
		})
	}
}

func Test_playlistsIDsSlice(t *testing.T) {
	type args struct {
		playlists []*youtube.Playlist
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Sample 1",
			args: args{playlists: []*youtube.Playlist{{Id: "1"}, {Id: "2"}, {Id: "3"}}},
			want: []string{"1", "2", "3"},
		},
		{
			name: "Empty",
			args: args{playlists: []*youtube.Playlist{}},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := playlistsIDsSlice(tt.args.playlists); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("playlistsIDsSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_joinPlaylistDeletionPrefix(t *testing.T) {
	type args struct {
		playlistIndex int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Sample 1",
			args: args{playlistIndex: 123},
			want: "delete_123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinPlaylistDeletionPrefix(tt.args.playlistIndex); got != tt.want {
				t.Errorf("joinPlaylistDeletionPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateAuthLink(t *testing.T) {
	type args struct {
		generator authCodeURLGenerator
		state     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Sample",
			args: args{
				generator: &configMockT{url: "https://auth.example.com/auth"},
				state:     "123456789",
			},
			want: "https://auth.example.com/auth?state=123456789",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAuthLink(tt.args.generator, tt.args.state); got != tt.want {
				t.Errorf("generateAuthLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOAuthStateAndToken(t *testing.T) {
	type args struct {
		c formValueGetter
	}
	tests := []struct {
		name string
		args args
		want *oauthRequestData
	}{
		{
			name: "Sample",
			args: args{c: newFormValueGetterMockT(map[string]string{
				oauthGoogleGETVariableCode:  "code123456789",
				oauthGoogleGETVariableState: "state123456789",
			})},
			want: &oauthRequestData{
				State: "state123456789",
				Code:  "code123456789",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOAuthStateAndToken(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOAuthStateAndToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_countItemsOfPlaylists(t *testing.T) {
	type args struct {
		playlists []*youtube.Playlist
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Sample 1",
			args: args{playlists: []*youtube.Playlist{
				{ContentDetails: &youtube.PlaylistContentDetails{ItemCount: 15}},
				{ContentDetails: &youtube.PlaylistContentDetails{ItemCount: 7}},
				{},
				{ContentDetails: &youtube.PlaylistContentDetails{ItemCount: 3}},
				{ContentDetails: &youtube.PlaylistContentDetails{ItemCount: 0}},
				{},
			}},
			want: 25,
		},
		{
			name: "Sample 2",
			args: args{playlists: []*youtube.Playlist{}},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countItemsOfPlaylists(tt.args.playlists); got != tt.want {
				t.Errorf("countItemsOfPlaylists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateState(t *testing.T) {
	rand.Seed(0)
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Sample",
			want: "18a8de5376e5daf6c6ce06df8f6e5bc1f8108e7a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateState(); got != tt.want {
				t.Errorf("generateState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPlaylistsIDsFromLinks(t *testing.T) {
	type args struct {
		links []string
	}
	tests := []struct {
		name             string
		args             args
		want             []string
		wantErrorsNumber int
	}{
		{
			name: "Sample 1",
			args: args{links: []string{
				"https://www.youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-1",
				"   https://www.youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-2   ",
				"   https://youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-3   ",
				"https://www.youtube.com/playlist",
				"https://ru.wikipedia.org/wiki?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-4",
				"",
				"https://www.youtube.com/playlist?list=PLTVdmvDFrwP\x00hfnZPXCdUkvy-EH3GnFa-1",
			}},
			want: []string{
				"PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-1",
				"PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-2",
				"PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-3",
			},
			wantErrorsNumber: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errors := playlistsIDsFromLinks(tt.args.links)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("playlistsIDsFromLinks() -> %v", diff)
				return
			}
			if len(errors) != tt.wantErrorsNumber {
				t.Errorf("playlistsIDsFromLinks() got errors = %v, want %v", len(errors), tt.wantErrorsNumber)
			}
		})
	}
}
