// Copyright © 2016 NAME HERE <joe@5sigma.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"conduit/engine"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var executeCmd = &cobra.Command{
	Use:   "execute [script]",
	Short: "Execute a script file.",
	Long:  `Execute a Javascript file at the specified path.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("slave:", cmd.Flag("slave").Value.String())
		fmt.Println("port:", cmd.Flag("port").Value.String())
		if cmd.Flag("slave").Value.String() == "true" {
			fmt.Println("Running in slave mode.")
		}
		if len(args) == 0 {
			fmt.Println("No file specified.")
			os.Exit(-1)
		}
		startTime := time.Now()
		eng := engine.New()
		eng.ExecuteFile(args[0])
		fmt.Printf("Script executed in %s .\n", time.Since(startTime))
	},
}

func init() {
	RootCmd.AddCommand(executeCmd)
}
