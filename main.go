package main

import (
	"os"
	"regexp"

	"github.com/rancher/rke/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var VERSION = "v0.1.19-ske-1"
var released = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+-ske(-[0-9]+)?$`)

func main() {
	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	app := cli.NewApp()
	app.Name = "ske"
	app.Version = VERSION
	app.Usage = "Suranaree Kubernetes Engine, built on the RKE's technology"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		logrus.Debugf("SKE version %s", app.Version)
		if released.MatchString(app.Version) {
			return nil
		}
		logrus.Warnf("This is not an officially supported version (%s) of SKE. Please download the latest official release at https://github.com/chanwit/ske/releases/latest", app.Version)
		return nil
	}
	app.Author = "SKE Authors"
	app.Email = ""
	app.Commands = []cli.Command{
		cmd.UpCommand(),
		cmd.RemoveCommand(),
		cmd.VersionCommand(),
		cmd.ConfigCommand(),
		cmd.EtcdCommand(),
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug logging",
		},
	}
	return app.Run(os.Args)
}
