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
import ("errors")

type CloudType string 

const(
	Vsphere CloudType = "vsphere"
	Aws 		  	  = "aws"
	Azure             = "azure"
	//gcp TBD
)

type Credentials struct {
	Username    string
	Password    string 
	Private_key string 
}

type AppInstanceType struct {
	Infra       string 
    Cloudtype	CloudType
    Region      string 
	ResourceGrp string 
	Tags        []string
	Storage     []int
    Creds   	Credentials
	Config      string 
	Facts 		map[string]string
}

type StackType struct {
	Id           string              // must be unique per live session
	AppInstances map[string]*AppInstanceType  
}

func IsValidCloudType(t CloudType) error {
    switch t {
	case Azure:
        return nil
    }
    return errors.New("Invalid cloud type specified. Must be azure")
}


// global data structure
var StackInstance *StackType
