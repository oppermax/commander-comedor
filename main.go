package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	var cli CLI

	ctx := kong.Parse(&cli,
		kong.Name("commander-comedor"),
		kong.Description("Get the menu of the UGR comedores universitarios."),
		kong.UsageOnError(),
	)

	err := ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
