package cmd

import (
	"github.com/maxsid/playlists-copy/cli"
	"github.com/maxsid/playlists-copy/youtube/auth"
	"github.com/maxsid/playlists-copy/youtube/service"
	"github.com/spf13/cobra"
)

var cliCMD = &cobra.Command{
	Use:   "cli",
	Short: "Run program in CLI mode.",
	Run: func(cmd *cobra.Command, args []string) {
		cred, err := auth.LoadCredentialFromFile(credentialPath)
		if err != nil {
			panic(err)
		}
		cli.Run(userConfigDir, cred, service.NewYouTubeService())
	},
}
