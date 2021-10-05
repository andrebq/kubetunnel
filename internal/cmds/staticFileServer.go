package cmds

import (
	"net/http"

	"github.com/urfave/cli/v2"
)

func StaticFileServer() *cli.Command {
	var bindAddr string
	var directory string
	return &cli.Command{
		Name:  "static-file-server",
		Usage: "A dead simple Go static file server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to listen for incoming connections",
				EnvVars:     []string{"KUBETUNNEL_STATIC_FILE_SERVER_BIND"},
				Required:    true,
				Destination: &bindAddr,
			},
			&cli.StringFlag{
				Name:        "directory",
				Aliases:     []string{"dir"},
				Usage:       "Directory with the static content",
				EnvVars:     []string{"KUBETUNNEL_STATIC_FILE_SERVER_DIRECTORY"},
				Required:    true,
				Destination: &directory,
			},
		},
		Action: func(c *cli.Context) error {
			return http.ListenAndServe(bindAddr, http.FileServer(http.Dir(directory)))
		},
	}
}
