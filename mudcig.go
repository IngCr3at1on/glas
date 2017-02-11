package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/IngCr3at1on/mudcig/core"
)

var (
	cfgFile  string
	charFile string

	cmd = &cobra.Command{
		Use:   "mudcig",
		Short: "A simple MUD Client In Go",

		Run: func(cmd *cobra.Command, args []string) {
			core.Start(args[0])
		},
	}
)

func init() {
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mudcig.yaml)")
	cmd.Flags().StringVarP(&charFile, "charfile", "c", "", "define a character file to start with")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("..mudcig")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
