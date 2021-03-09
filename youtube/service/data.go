package service

import (
	"context"
	"fmt"
	"github.com/maxsid/playlists-copy/youtube"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

type youTubeUserServiceCreator struct{}

func NewYouTubeServiceCreator() youtube.ServiceCreator {
	return &youTubeUserServiceCreator{}
}

func (y *youTubeUserServiceCreator) NewUserService() youtube.Service {
	return NewYouTubeService()
}

type playlistsListCallHandler func(call *youtubeAPI.PlaylistsListCall) *youtubeAPI.PlaylistsListCall

type youTubeUserService struct {
	part      []string
	maxResult int64
	service   *youtubeAPI.Service
}

func NewYouTubeService() youtube.Service {
	return &youTubeUserService{part: []string{"snippet", "id", "contentDetails"}, maxResult: 50}
}

func (y *youTubeUserService) ConfigUserService(ctx context.Context, config youtube.Config, token *oauth2.Token) (err error) {
	y.service, err = youtubeAPI.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	return
}

func (y *youTubeUserService) ChannelOfMine(ctx context.Context) (*youtubeAPI.Channel, error) {
	call := y.service.Channels.List(y.part).Context(ctx).Mine(true)
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("%w channel: mine", ErrNotFound)
	}
	return resp.Items[0], nil
}

func (y *youTubeUserService) playlistsList(ctx context.Context, callHandler playlistsListCallHandler) ([]*youtubeAPI.Playlist, error) {
	playlists := make([]*youtubeAPI.Playlist, 0)
	call := y.service.Playlists.List(y.part).Context(ctx).MaxResults(y.maxResult)
	if callHandler != nil {
		call = callHandler(call)
	}
	err := call.Pages(ctx, func(resp *youtubeAPI.PlaylistListResponse) error {
		playlists = append(playlists, resp.Items...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

func (y *youTubeUserService) PlaylistsOfChannel(ctx context.Context, channelID string) ([]*youtubeAPI.Playlist, error) {
	return y.playlistsList(ctx, func(call *youtubeAPI.PlaylistsListCall) *youtubeAPI.PlaylistsListCall {
		return call.ChannelId(channelID)
	})
}

func (y *youTubeUserService) PlaylistsByIDs(ctx context.Context, id ...string) ([]*youtubeAPI.Playlist, error) {
	return y.playlistsList(ctx, func(call *youtubeAPI.PlaylistsListCall) *youtubeAPI.PlaylistsListCall {
		return call.Id(id...)
	})
}

func (y *youTubeUserService) PlaylistByID(ctx context.Context, id string) (*youtubeAPI.Playlist, error) {
	ps, err := y.PlaylistsByIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	if ps == nil || len(ps) == 0 {
		return nil, fmt.Errorf("%w playlist: id %s", ErrNotFound, id)
	}
	return ps[0], nil
}

func (y *youTubeUserService) PlaylistItemsOfSeveralPlaylists(ctx context.Context, playlistID ...string) ([]*youtubeAPI.PlaylistItem, error) {
	items := make([]*youtubeAPI.PlaylistItem, 0)
	for _, pID := range playlistID {
		call := y.service.PlaylistItems.List(y.part).Context(ctx).PlaylistId(pID).MaxResults(y.maxResult)
		err := call.Pages(ctx, func(resp *youtubeAPI.PlaylistItemListResponse) error {
			items = append(items, resp.Items...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (y *youTubeUserService) InsertPlaylistItems(ctx context.Context, playlistID string, item ...*youtubeAPI.PlaylistItem) ([]*youtubeAPI.PlaylistItem, error) {
	for _, it := range item {
		newItem := new(youtubeAPI.PlaylistItem)
		newItem.Snippet = it.Snippet
		it.Snippet.PlaylistId = playlistID
		call := y.service.PlaylistItems.Insert(y.part, it).Context(ctx)
		if _, err := call.Do(); err != nil {
			return nil, err
		}
	}
	return item, nil
}
