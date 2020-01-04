package main

import (
	"log"
	"os"

	charter "github.com/maybetheresloop/charter-go"
	_ "github.com/maybetheresloop/charter-go/passwd/backend/text"
	"github.com/urfave/cli"
)

func run(ctx *cli.Context) error {
	// Get default configuration.
	conf := defaultConfig()

	// Next, retrieve configuration from configuration file, if specified.
	fp := ctx.Args().First()
	if fp != "" {
		if err := populateFromFile(conf, fp); err != nil {
			return err
		}
	}

	// Finally, retrieve configuration overrides from command line.
	populateFromCliContext(conf, ctx)

	srv := charter.NewServer(conf)

	return srv.ListenAndServe()
}

func main() {
	app := cli.NewApp()
	app.Name = "charterd"
	app.Description = "Start the Charter File Transfer Protocol server."
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("%s: %v\n", os.Args[0], err)
	}
}
