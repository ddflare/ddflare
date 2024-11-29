/*
Copyright © 2024 Francesco Giudici <dev@foggy.day>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

// This is called by main.main().
func Execute() {
	app := &cli.App{
		Usage: "update DNS entries via cloudflare APIs",
		Commands: []*cli.Command{
			newGetCommand(),
			newSetCommand(),
			newVersionCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "loglevel",
				Aliases: []string{"log"},
				Usage: "set the log level [" + strings.Join([]string{
					slog.LevelDebug.String(),
					slog.LevelInfo.String(),
					slog.LevelWarn.String(),
					slog.LevelError.String(),
				}, ",") + "]",
				EnvVars: []string{"LOGLEVEL"},
				Value:   slog.LevelInfo.String(),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output (shorthand for '--log DEBUG')",
				Value:   false,
			},
		},
		Before: func(cCtx *cli.Context) error {
			loglevel := cCtx.String("loglevel")
			verbose := cCtx.Bool("verbose")
			if verbose {
				loglevel = slog.LevelDebug.String()
			}

			var slogLvl slog.Level
			switch loglevel {
			case slog.LevelDebug.String():
				slogLvl = slog.LevelDebug
			case slog.LevelInfo.String():
				slogLvl = slog.LevelInfo
			case slog.LevelWarn.String():
				slogLvl = slog.LevelWarn
			case slog.LevelError.String():
				slogLvl = slog.LevelError
			default:
				return fmt.Errorf("unknown log level: %s", loglevel)
			}
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLvl}))
			slog.SetDefault(logger)
			slog.Debug("logging started", "Log Level", slogLvl.String())
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
