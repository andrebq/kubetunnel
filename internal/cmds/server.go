package cmds

import (
	"github.com/andrebq/kubetunnel"
	"github.com/urfave/cli/v2"
)

func Server() *cli.Command {
	var serverBind string
	return &cli.Command{
		Name:  "server",
		Usage: "Open the tunnel server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				EnvVars:     []string{"KUBETUNNEL_SERVER_BIND"},
				Usage:       "Address to listen for incoming tunnel connections",
				Destination: &serverBind,
			},
		},
		Action: func(c *cli.Context) error {
			s := kubetunnel.NewServer()
			return s.Run(c.Context, serverBind)
		},
	}
}
