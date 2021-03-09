package cmd

import (
	"github.com/maxsid/playlists-copy/server"
	"github.com/maxsid/playlists-copy/youtube/auth"
	"github.com/maxsid/playlists-copy/youtube/service"
	"github.com/spf13/cobra"
)

var (
	serverAddress = ":8080"
)

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "Run web server",
	Run: func(cmd *cobra.Command, args []string) {
		cred, err := auth.LoadCredentialFromFile(credentialPath)
		if err != nil {
			panic(err)
		}
		server.Run(serverAddress, cred, service.NewYouTubeServiceCreator())
	},
}

func initServerFlags() {
	serverCMD.PersistentFlags().StringVar(&serverAddress, "addr", serverAddress, "Server listening address")
}
