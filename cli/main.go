package cli

import (
	"bufio"
	"context"
	"fmt"
	"github.com/maxsid/playlists-copy/youtube"
	"github.com/maxsid/playlists-copy/youtube/helper"
	youtubeAPI "google.golang.org/api/youtube/v3"
	"log"
	"os"
)

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func readDestinationPlaylist(playlistGetter youtube.ServicePlaylistsGetter, channel *youtubeAPI.Channel) (*youtubeAPI.Playlist, error) {
	var destUrl string
	fmt.Print("Enter destination playlist url: ")
	_, _ = fmt.Scan(&destUrl)
	destinationPlaylistID, err := helper.YoutubePlaylistIDFromURL(destUrl)
	if err != nil {
		return nil, err
	}
	playlist, err := playlistGetter.PlaylistByID(context.TODO(), destinationPlaylistID)
	if err != nil {
		return nil, err
	}
	if playlist.Snippet.ChannelId != channel.Id {
		return nil, fmt.Errorf("playlist %s is not yours", destinationPlaylistID)
	}
	return playlist, nil
}

func readSourcePlaylists(getter youtube.ServicePlaylistsGetter) ([]*youtubeAPI.Playlist, error) {
	playlistsIDs := make([]string, 0)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter source playlists (empty line for stop): ")
	for scanner.Scan() {
		rawURL := scanner.Text()
		if rawURL == "" {
			break
		}
		id, err := helper.YoutubePlaylistIDFromURL(rawURL)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		playlistsIDs = append(playlistsIDs, id)
	}
	ps, err := getter.PlaylistsByIDs(context.TODO(), playlistsIDs...)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func mapPlaylistsIDs(playlists []*youtubeAPI.Playlist) []string {
	ids := make([]string, len(playlists))
	for i, v := range playlists {
		ids[i] = v.Id
	}
	return ids
}

func Run(configDir string, credential youtube.Config, manager youtube.Service) {
	err := setService(context.TODO(), manager, credential, configDir)
	handleError(err, "")

	myChannel, err := manager.ChannelOfMine(context.TODO())
	handleError(err, "")
	log.Printf("Your channel is %s (id: %s)", myChannel.Snippet.Title, myChannel.Id)

	myPlaylist, err := readDestinationPlaylist(manager, myChannel)
	handleError(err, "")
	log.Printf("Selected %s (id: %s) playlist", myPlaylist.Snippet.Title, myChannel.Id)

	sourcePlaylists, err := readSourcePlaylists(manager)
	handleError(err, "")
	log.Printf("Selected %d playlists", len(sourcePlaylists))

	items, err := manager.PlaylistItemsOfSeveralPlaylists(context.TODO(), mapPlaylistsIDs(sourcePlaylists)...)
	handleError(err, "")
	log.Printf("Found %d videos", len(items))

	log.Printf("Start inserting")
	_, err = manager.InsertPlaylistItems(context.TODO(), myPlaylist.Id, items...)
	handleError(err, "")
}
