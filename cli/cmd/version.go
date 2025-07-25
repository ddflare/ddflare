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

	"github.com/urfave/cli/v2"

	"github.com/ddflare/ddflare/pkg/version"
)

func newVersionCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "version",
		Usage: "print the version and exit",
		Action: func(cCtx *cli.Context) error {
			fmt.Println(version.Version)
			return nil
		},
	}
	return cmd
}
