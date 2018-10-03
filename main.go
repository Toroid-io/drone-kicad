package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var build = "0" // build number set at compile-time

func main() {

	app := cli.NewApp()
	app.Name = "kicad plugin"
	app.Usage = "kicad plugin"
	app.Action = run
	app.Version = fmt.Sprintf("0.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "client.code",
			Usage:  "enterprise client code",
			EnvVar: "PLUGIN_CLIENT_CODE",
		},
		cli.StringFlag{
			Name:   "client.name",
			Usage:  "client name",
			EnvVar: "PLUGIN_CLIENT_NAME",
		},
		cli.StringFlag{
			Name:   "projects",
			Usage:  "projects structure",
			EnvVar: "PLUGIN_PROJECTS",
		},
		cli.StringFlag{
			Name:   "deps",
			Usage:  "Download dependencies",
			EnvVar: "PLUGIN_DEPENDENCIES",
		},
		cli.StringFlag{
			Name:   "netrc.machine",
			Usage:  "netrc machine",
			EnvVar: "DRONE_NETRC_MACHINE",
		},
		cli.StringFlag{
			Name:   "netrc.username",
			Usage:  "netrc username",
			EnvVar: "DRONE_NETRC_USERNAME",
		},
		cli.StringFlag{
			Name:   "netrc.password",
			Usage:  "netrc password",
			EnvVar: "DRONE_NETRC_PASSWORD",
		},
		cli.StringFlag{
			Name:   "commit.tag",
			Usage:  "commit tag",
			EnvVar: "DRONE_TAG",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {

	plugin := Plugin{
		Netrc: Netrc{
			Login:    c.String("netrc.username"),
			Machine:  c.String("netrc.machine"),
			Password: c.String("netrc.password"),
		},
		Commit: Commit{
			Tag: c.String("commit.tag"),
			Sha: c.String("commit.sha"),
		},
	}

	err := json.Unmarshal([]byte(c.String("deps")), &plugin.Dependencies)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(c.String("projects")), &plugin.Projects)
	if err != nil {
		return err
	}

	return plugin.Exec()
}
