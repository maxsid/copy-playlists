package youtube

import (
	"context"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

type Service interface {
	userServiceConfigurator
	ServiceChannelsGetter
	ServicePlaylistsGetter
	playlistItemsGetter
	playlistItemsInserter
}

type ServiceCreator interface {
	NewUserService() Service
}

type userServiceConfigurator interface {
	ConfigUserService(ctx context.Context, conf Config, token *oauth2.Token) error
}

type ServiceChannelsGetter interface {
	userServiceConfigurator
	ChannelOfMine(ctx context.Context) (*youtube.Channel, error)
}

type ServicePlaylistsGetter interface {
	userServiceConfigurator
	PlaylistsByIDs(ctx context.Context, id ...string) ([]*youtube.Playlist, error)
	PlaylistByID(ctx context.Context, id string) (*youtube.Playlist, error)
	PlaylistsOfChannel(ctx context.Context, channelID string) ([]*youtube.Playlist, error)
}

type playlistItemsGetter interface {
	userServiceConfigurator
	PlaylistItemsOfSeveralPlaylists(ctx context.Context, playlistID ...string) ([]*youtube.PlaylistItem, error)
}

type playlistItemsInserter interface {
	userServiceConfigurator
	InsertPlaylistItems(ctx context.Context, playlistID string, item ...*youtube.PlaylistItem) ([]*youtube.PlaylistItem, error)
}
