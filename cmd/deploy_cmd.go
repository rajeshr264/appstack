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
	"fmt"
	"os"
	"strings"

	"rajeshr264/ephstack/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		// pass the stack file name to the populate the data structures & deploy
		if err := parse(args[0]); err != nil {
			cobra.CheckErr(err)
		}
	},
}

func parse(stackFileName string) error {
	viper.SetConfigFile(stackFileName)

	// If a config file is found, read it in.
	yamlData, _ := os.Open(stackFileName)
	if err := viper.ReadConfig(yamlData); err != nil {
		return errors.New("Unable to parse " + stackFileName)
	}

	var stackInstance =  new(ephstack.StackType)
	fmt.Fprintln(os.Stderr, "Reading stack file:", viper.ConfigFileUsed())
	stackInstance.Id = viper.GetString("stack.name")

	// walk thru the 'apps' decls
	vAppInstancesTree := viper.Sub("stack.apps")
	if vAppInstancesTree == nil { // Sub returns nil if the key cannot be found
		return errors.New("Malformed YAML file. Unable to parse " + stackFileName)
	}
	// populate the StackInstance.appInstances Hash
	var appInstMap map[string]*(ephstack.AppInstanceType) = make(map[string]*(ephstack.AppInstanceType))
	
	for _, keyVal := range vAppInstancesTree.AllKeys() {
		var keyValPair []string = strings.Split(keyVal, ".") // viper returns "app1.config"
		var appInst *ephstack.AppInstanceType = appInstMap[keyValPair[0]]
		if appInst == nil { // uninitalized
			appInst = &ephstack.AppInstanceType{
				Infra:       "",
				Cloudtype:   "",
				Region:      "",
				ResourceGrp: "",
				Tags:        make([]string, 0),
				Storage:     make([]int, 0),
				Creds:       ephstack.Credentials{Username: "", Password: "", Private_key: ""},
				Config:      "",
                Facts:       make(map[string]string),
			}
		}
		switch keyValPair[1] {
		case "config":
			appInst.Config = vAppInstancesTree.Get(keyVal).(string)
			appInstMap[keyValPair[0]] = appInst
		case "infra":
			appInst.Infra = vAppInstancesTree.Get(keyVal).(string)
			appInstMap[keyValPair[0]] = appInst
		case "tags":
			appInst.Tags = append(appInst.Tags, vAppInstancesTree.Get(keyVal).(string))
			appInstMap[keyValPair[0]] = appInst
		case "facts":
            var mapString map[string]string = make(map[string]string) 
            for key,value := range vAppInstancesTree.Get(keyVal).([]interface{}) {
                var strKey string   = fmt.Sprintf("%v", key)
                var strValue string = fmt.Sprintf("%v", value)
                mapString[strKey] = strValue
            }
            appInst.Facts = mapString
            appInstMap[keyValPair[0]] = appInst
		default:
            err1 := errors.New("Unexpected App instance value '" + keyValPair[1] + "' found in " + stackFileName)
			return err1 
		}
	}

	// parse the cloud config files 
	


	stackInstance.AppInstances = appInstMap
	ephstack.StackInstance = stackInstance
	return nil
}


func init() {

	// define your flags and configuration settings.
	rootCmd.AddCommand(deployCmd)

	// read in the stack file first 
	
	// parse the config files and only read in the settings for the 
	// cloud configs specified in the stack file

}
