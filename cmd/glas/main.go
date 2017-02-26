package main

import (
	"fmt"
	"os"

	"io"

	"github.com/IngCr3at1on/glas"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	charFile string
	address  string

	iochan chan string
	ioout  io.Writer
	ioerr  io.Writer

	cmd = &cobra.Command{
		Use:   "glas",
		Short: "A simple MUD Client In Go",

		Run: func(cmd *cobra.Command, args []string) {
			glas.Start(iochan, ioout, ioerr, charFile, address)
		},
	}
)

func init() {
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.glas.yaml)")
	cmd.Flags().StringVarP(&charFile, "charfile", "c", "", "define a character file to start with")
	cmd.Flags().StringVarP(&address, "address", "a", "", "mud connection address")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("..glas")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func errAndExit(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(-1)
}

func main() {
	iochan = make(chan string)
	ioout = os.Stdout
	ioerr = os.Stderr

	go handleInput()
	errAndExit(cmd.Execute())
}
