package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path"

	"github.com/spf13/viper"
)

var (
	credentialPath string
	userConfigDir  string

	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "playlists-copy",
	Short: "Copies public playlists into one.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(cliCMD)
	rootCmd.AddCommand(serverCMD)

	initRootFlags()
	initServerFlags()
}

func initRootFlags() {
	cfgDir, err := getConfigDirectory()
	if err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", fmt.Sprintf("%s/config.yaml", cfgDir), "config file")
	rootCmd.PersistentFlags().StringVarP(&credentialPath, "credential", "c", "", "(required) a json credential file from Google Cloud Console")
	if err := rootCmd.MarkPersistentFlagRequired("credential"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory.
		dir, err := getConfigDirectory()
		cobra.CheckErr(err)
		userConfigDir = dir

		// Search config in dir directory with name "config" (without extension).
		viper.AddConfigPath(dir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func getConfigDirectory() (string, error) {
	if userConfigDir != "" {
		return userConfigDir, nil
	}
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(configPath, "playlists-copy"), nil
}
