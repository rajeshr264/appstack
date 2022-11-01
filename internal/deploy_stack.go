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
package ephstack

import (
	"errors"
	"fmt"
	"os"
	//"reflect"
	"strings"

	"github.com/spf13/viper"
)

// global data structure
var stackInstance StackType = StackType{id: "", appInstances: nil}

func DeployStack(stackFileName string) error {
	viper.SetConfigFile(stackFileName)

	// If a config file is found, read it in.
	yamlData, _ := os.Open(stackFileName)
	if err := viper.ReadConfig(yamlData); err != nil {
		return errors.New("Unable to parse " + stackFileName)
	}

	fmt.Fprintln(os.Stderr, "Reading stack file:", viper.ConfigFileUsed())
	stackInstance.id = viper.GetString("stack.name")

	// walk thru the 'apps' decls
	vAppInstancesTree := viper.Sub("stack.apps")
	if vAppInstancesTree == nil { // Sub returns nil if the key cannot be found
		return errors.New("Malformed YAML file. Unable to parse " + stackFileName)
	}
	// populate the stackInstance.appInstances Hash
	var appInstMap map[string]*AppInstanceType = make(map[string]*AppInstanceType)
	for _, keyVal := range vAppInstancesTree.AllKeys() {
		var keyValPair []string = strings.Split(keyVal, ".") // viper returns "app1.config"
		var appInst *AppInstanceType = appInstMap[keyValPair[0]]
		if appInst == nil { // uninitalized
			appInst = &AppInstanceType{
				infra:       "",
				cloudtype:   "",
				region:      "",
				resourceGrp: "",
				tags:        make([]string, 0),
				storage:     make([]int, 0),
				creds:       Credentials{username: "", password: "", private_key: ""},
				config:      "",
                facts:       make(map[string]string),
			}
		}
		switch keyValPair[1] {
		case "config":
			appInst.config = vAppInstancesTree.Get(keyVal).(string)
			appInstMap[keyValPair[0]] = appInst
		case "infra":
			appInst.infra = vAppInstancesTree.Get(keyVal).(string)
			appInstMap[keyValPair[0]] = appInst
		case "tags":
			appInst.tags = append(appInst.tags, vAppInstancesTree.Get(keyVal).(string))
			appInstMap[keyValPair[0]] = appInst
		case "facts":
            var mapString map[string]string = make(map[string]string) 
            for key,value := range vAppInstancesTree.Get(keyVal).([]interface{}) {
                var strKey string   = fmt.Sprintf("%v", key)
                var strValue string = fmt.Sprintf("%v", value)
                mapString[strKey] = strValue
            }
            appInst.facts = mapString
            appInstMap[keyValPair[0]] = appInst
		default:
			//errors.New("Unexpected App instance value '" + keyValPair[1] + "' found in " + stackFileName)
		}
	}
	stackInstance.appInstances = appInstMap
	if len(stackInstance.appInstances) == 0 {
		fmt.Println("map is empty")
	}
	for key, element := range stackInstance.appInstances {
		fmt.Println("Key:", key, "=>", "Element:", element)
	}
	return nil
}
