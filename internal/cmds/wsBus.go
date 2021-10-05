package cmds

import (
	"github.com/andrebq/kubetunnel/internal/demo/wsbus"
	"github.com/urfave/cli/v2"
)

func WebSocketBus() *cli.Command {
	var bindAddr, directory string
	return &cli.Command{
		Name:  "websocket-bus",
		Usage: "Creates a Websocket bus which broadcasts any message to all clients",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to listen for new connections",
				EnvVars:     []string{"KUBETUNNEL_WEBSOCKET_BUS_BIND"},
				Required:    true,
				Destination: &bindAddr,
			},
			&cli.StringFlag{
				Name:        "directory",
				Usage:       "Directory to serve static files",
				EnvVars:     []string{"KUBETUNNEL_WEBSOCKET_BUS_DIRECTORY"},
				Required:    true,
				Destination: &directory,
			},
		},
		Action: func(c *cli.Context) error {
			return wsbus.Run(bindAddr, directory)
		},
	}
}
