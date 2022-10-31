/*
Copyright © 2022 Rajesh Radhakrishnan enthoughts@gmail.com

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
	"errors"
	"os"

	"github.com/spf13/cobra"
	//"rajeshr264/ephstack/internal"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy <stack file>",
	Short: "Deploy the app(s) specified in the stack file",

	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := errors.New("<stack file> was not specified")
			cobra.CheckErr(err)
		}
		stackFile := args[0]
		// read in the stack file.
		_, err := os.Open(stackFile)
		cobra.CheckErr(err)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// pass the stack file name to the populate the data structures
		// err = ephstack.DeployStack(args[0])
		// cobra.CheckErr(err)
	},
}

func init() {

	// define your flags and configuration settings.
	rootCmd.AddCommand(deployCmd)

	// read in the stack file first 
	
	// parse the config files and only read in the settings for the 
	// cloud configs specified in the stack file

}
