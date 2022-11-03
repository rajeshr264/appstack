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
package main

import (
	"fmt"

	"rajeshr264/ephstack/cmd"
    "rajeshr264/ephstack/internal"
)

func main() {
	cmd.Execute()
	
	if len(ephstack.StackInstance.AppInstances) == 0 {
		fmt.Println("map is empty")
	}
	for key, element := range ephstack.StackInstance.AppInstances {
		fmt.Println("Key:", key, "=>", "Element:", *element)
	}
	
	var ihw ephstack.InfraHWInstancesMapType = *ephstack.InfraHWInstances
	for cloudName, cloudStacks := range ihw {
		fmt.Println("Cloud: " + cloudName)
		for k,v := range (*cloudStacks) {
			fmt.Print("k = ")
			fmt.Print(k)
			fmt.Print(" v = ")
			fmt.Println(v)
		}
	}	
}
