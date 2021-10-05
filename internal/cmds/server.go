package cmds

import (
	"github.com/andrebq/kubetunnel"
	"github.com/urfave/cli/v2"
)

func Server() *cli.Command {
	var serverBind string
	var targetBind string
	return &cli.Command{
		Name:  "server",
		Usage: "Open the tunnel server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				EnvVars:     []string{"KUBETUNNEL_SERVER_BIND"},
				Usage:       "Address to listen for incoming tunnel connections",
				Destination: &serverBind,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "target",
				EnvVars:     []string{"KUBETUNNEL_SERVER_TARGET"},
				Usage:       "Address on which the server will accept incoming requests to proxy via the Websocket connections on bind",
				Destination: &targetBind,
				Required:    true,
			},
		},
		Action: func(c *cli.Context) error {
			s := kubetunnel.NewServer()
			return s.Run(c.Context, serverBind, targetBind)
		},
	}
}
