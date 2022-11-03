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
	"path/filepath"
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

func parseStackFile(stackFileName string) error {

	// set the stack file that viper needs to parse 
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
	var appInstances map[string]*(ephstack.AppInstanceType) = make(map[string]*(ephstack.AppInstanceType))
	
	for _, keyVal := range vAppInstancesTree.AllKeys() {
		var keyValPair []string = strings.Split(keyVal, ".") // viper returns "app1.config"
		var appInst *ephstack.AppInstanceType = appInstances[keyValPair[0]]
		if appInst == nil { // uninitalized
			appInst = &ephstack.AppInstanceType{
				Infra:       "",
				Creds:       ephstack.Credentials{Username: "", Password: "", Private_key: ""},
				Config:      "",
                Facts:       make(map[string]string),
			}
		}
		switch keyValPair[1] {
		case "config":
			appInst.Config = vAppInstancesTree.Get(keyVal).(string)
			appInstances[keyValPair[0]] = appInst
		case "infra":
			appInst.Infra = vAppInstancesTree.Get(keyVal).(string)
			appInstances[keyValPair[0]] = appInst
		case "facts":
            var factsMap map[string]string = make(map[string]string) 
            for key,value := range vAppInstancesTree.Get(keyVal).([]interface{}) {
                var strKey string   = fmt.Sprintf("%v", key)
                var strValue string = fmt.Sprintf("%v", value)
                factsMap[strKey] = strValue
            }
            appInst.Facts = factsMap
            appInstances[keyValPair[0]] = appInst
		default:
            err := errors.New("unexpected App instance value '" + keyValPair[1] + "' found in " + stackFileName)
			return err
		}
	}

	// Save the stack info
	stackInstance.AppInstances = appInstances
	ephstack.StackInstance = stackInstance
	return nil 
}

func parseConfigFiles() error {

	// parse all the config files 
	var configFileNames []string 
	err := filepath.Walk("config", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			err1 := errors.New("unable to open directory ./config")
			return err1
		}
		info, err2 := os.Stat(path)
		if info.IsDir() {} else {
			configFileNames = append(configFileNames,path)
		} 
		if err2 != nil { 
			return err2
		}
		return nil
	})
	if err != nil {
		err1 := errors.New("unable to open sub directory 'config'")
		return err1
	}

	// a map to store all the cloud infra settings 
	infraHWInstancesMap := &ephstack.InfraHWInstancesMapType{}

	for _,configFileName := range configFileNames {
		viper.SetConfigFile(configFileName)
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil { // Handle errors reading the config file
			err1 := errors.New("unable to read config file" + configFileName)
			return err1
		}

		vConfigTree := viper.Sub("config.infra")
		if vConfigTree == nil {
			err = errors.New("config file: " + configFileName + " YAML syntax is not correct")
			return err
		}
		
		// per cloud map 
		infraHWMap := &ephstack.InfraHWInstMapType{}
		(*infraHWInstancesMap)[viper.GetString("config.cloud")] = infraHWMap
		
		for _, keyVal := range vConfigTree.AllKeys() {
			var keyValPair []string = strings.Split(keyVal, ".") // viper returns "<key>.size"

			// each cloud config stack is stored in InfraHwType  objects 
			var infraHWInst *ephstack.InfraHwType = (*infraHWMap)[keyValPair[0]]
		    if infraHWInst == nil { // uninitalized
		        infraHWInst = &ephstack.InfraHwType {
					Name       : "",
					Region     : "", 
					Type       : "",
					Image      : "",
					Disks      : make([]string,0),
		            Tags       : make(map[string]string,0),
				}
			}  

			switch keyValPair[1] {
			case "image":
			  infraHWInst.Image = vConfigTree.Get(keyVal).(string)
			  (*infraHWMap)[keyValPair[0]] = infraHWInst
			case "region":
				infraHWInst.Region  = vConfigTree.Get(keyVal).(string)
				(*infraHWMap)[keyValPair[0]] = infraHWInst
			case "type":
				infraHWInst.Type  = vConfigTree.Get(keyVal).(string)
				(*infraHWMap)[keyValPair[0]] = infraHWInst
			case "disk":
				for _,v := range vConfigTree.Get(keyVal).([]interface{}) {
					var value string = fmt.Sprintf("%v", v)
					infraHWInst.Disks = append(infraHWInst.Disks, value)
				}
				(*infraHWMap)[keyValPair[0]] =  infraHWInst
			case "tags":
				var tagMap map[string]string = make(map[string]string) 
           		for key,value := range vConfigTree.Get(keyVal).([]interface{}) {
            		var strKey string   = fmt.Sprintf("%v", key)
            		var strValue string = fmt.Sprintf("%v", value)
            		tagMap[strKey] = strValue
           		}
				infraHWInst.Tags             = tagMap
            	(*infraHWMap)[keyValPair[0]] = infraHWInst
			default:
				err := errors.New("unexpected Config instance value " + keyValPair[1] + " found in " + configFileName)
				return err
			}	
		}
		 
		ephstack.InfraHWInstances = infraHWInstancesMap
	}	

	return nil 
}

func parse(stackFileName string) error {

	err := parseStackFile(stackFileName)
	if err != nil {
		return err
	}

	err = parseConfigFiles()
	if err != nil {
		return err 
	}

	return nil
}


func init() {

	// define your flags and configuration settings.
	rootCmd.AddCommand(deployCmd)

	// read in the stack file first 
	
	// parse the config files and only read in the settings for the 
	// cloud configs specified in the stack file

}
