package cmd

import (
	"fmt"
	"github.com/5sigma/conduit/info"
	"github.com/5sigma/conduit/log"
	"github.com/kardianos/osext"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "conduit",
	Short: "Conduit v" + info.ConduitVersion,
	Long: `Conduit ` + info.ConduitVersion + `
Conduit is a client/server package that allows command and management
of resources using JavaScript based automation scripts.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file path.")
	RootCmd.PersistentFlags().BoolVarP(&log.ShowDebug, "debug", "d", false,
		"Print debug information")
}

func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	exePath, _ := osext.ExecutableFolder()
	wd, _ := os.Getwd()
	viper.SetConfigName("conduit") // name of config file (without extension)
	viper.AddConfigPath(wd)
	viper.AddConfigPath(exePath)
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
