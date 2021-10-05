package main

import (
	"context"
	"os"

	"github.com/andrebq/kubetunnel/internal/cmds"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "kubetunnel",
		Commands: []*cli.Command{
			cmds.Client(), cmds.Server(), cmds.StaticFileServer(),
		},
	}

	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		log.Error().Err(err).Msg("Application failed")
	}
}
