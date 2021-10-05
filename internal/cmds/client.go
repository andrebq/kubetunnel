package cmds

import (
	"github.com/andrebq/kubetunnel"
	"github.com/urfave/cli/v2"
)

func Client() *cli.Command {
	var remoteEndpoint, localEndpoint, serverAddress string
	return &cli.Command{
		Name:  "client",
		Usage: "Exposes a service running in the local host to the remote host",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "server-address",
				EnvVars:     []string{"KUBETUNNEL_CLIENT_SERVER_ADDRESS"},
				Aliases:     []string{"s", "server", "server-addr"},
				Usage:       "Address where the service is running",
				Required:    true,
				Destination: &serverAddress,
			},
			&cli.StringFlag{
				Name:        "local-endpoint",
				EnvVars:     []string{"KUBETUNNEL_CLIENT_LOCAL_ENDPOINT"},
				Aliases:     []string{"local"},
				Usage:       "Local endpoint which will receive the proxied connections",
				Required:    true,
				Destination: &localEndpoint,
			},
			&cli.StringFlag{
				Name:        "remote-endpoint",
				EnvVars:     []string{"KUBETUNNEL_CLIENT_REMOTE_ENDPOINT"},
				Aliases:     []string{"remote"},
				Usage:       "Address (use :<port> to ignore the domain) which will recieve connections on the remote server",
				Required:    true,
				Destination: &remoteEndpoint,
			},
		},

		Action: func(appCtx *cli.Context) error {
			cli, err := kubetunnel.NewClient(appCtx.Context, remoteEndpoint, localEndpoint, serverAddress)
			if err != nil {
				return err
			}
			return cli.Run(appCtx.Context)
		},
	}
}
