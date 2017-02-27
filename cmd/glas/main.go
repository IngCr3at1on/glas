package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/IngCr3at1on/glas"
	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	_main    = "main"
	_cmdline = "cmdline"
)

var (
	_quit chan struct{}

	cfgFile  string
	confFile string
	charFile string
	address  string

	_conf *conf

	iochan  chan string
	cmdLine *gocui.View
	ioout   *gocui.View
	ioerr   io.Writer

	cmd = &cobra.Command{
		Use:   "glas",
		Short: "A simple MUD Client In Go",

		Run: func(cmd *cobra.Command, args []string) {
			var err error
			_conf, err = loadConf(confFile)
			errAndExit(err)

			_quit = make(chan struct{})
			iochan = make(chan string)
			ioerr = os.Stderr

			gui, err := gocui.NewGui(gocui.Output256)
			errAndExit(err)
			defer gui.Close()

			gui.SetManagerFunc(layout)
			gui.Mouse = true
			gui.Cursor = true

			errAndExit(initKeybindings(gui))

			go func(gui *gocui.Gui) {
				errAndExit(gui.MainLoop())
			}(gui)

			time.Sleep(100 * time.Millisecond)

			go func(gui *gocui.Gui) {
				for {
					// TODO make this work properly with _quit ?
					time.Sleep(100 * time.Millisecond)
					select {
					case <-_quit:
						return
					default:
						gui.Execute(func(gui *gocui.Gui) error {
							return nil
						})
					}
				}
			}(gui)

			glas.Start(iochan, ioout, ioerr, charFile, address, _quit)
		},
	}
)

func init() {
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.glas.yaml)")
	cmd.Flags().StringVarP(&confFile, "settings", "s", "", "define a settings file for global client settings")
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

func layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()

	var err error
	ioout, err = gui.SetView(_main, -1, -1, maxX, maxY-5)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Wrap(err, "gui.SetView")
		}
	}

	ioout.Autoscroll = true
	ioout.Wrap = true

	_, err = gui.SetView(_cmdline, -1, maxY-5, maxX, maxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Wrap(err, "gui.SetView")
		}
	}
	cmdLine, err = gui.SetCurrentView(_cmdline)
	if err != nil {
		return errors.Wrap(err, "gui.SetCurrentView")
	}

	cmdLine.Wrap = true
	cmdLine.Editable = true

	cmdLine.MoveCursor(0, 0, true)
	return nil
}

func initKeybindings(gui *gocui.Gui) error {
	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		return errors.Wrap(err, "gui.SetKeybinding")
	}

	if err := gui.SetKeybinding(_main, gocui.MouseWheelUp, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		// TODO I'm not sure how to do this...
		return nil
	}); err != nil {
		return errors.Wrap(err, "gui.SetKeybinding")
	}
	if err := gui.SetKeybinding(_main, gocui.MouseWheelDown, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		// TODO same as abovg...
		return nil
	}); err != nil {
		return errors.Wrap(err, "gui.SetKeybinding")
	}

	if err := gui.SetKeybinding(_cmdline, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		// TODO text should be highlighted to allow for clearing with one key
		// when not cleared.
		if _conf.InputAutoErase {
			iochan <- view.Buffer()
			view.Clear()
			view.MoveCursor(0, 0, false)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "gui.SetKeybinding")
	}

	return nil
}

func errAndExit(err error) {
	if err == nil {
		return
	}
	if err == gocui.ErrQuit {
		close(_quit)
		return
	}
	fmt.Println(err.Error())
	os.Exit(-1)
}

func main() {
	errAndExit(cmd.Execute())
}
