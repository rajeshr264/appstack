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
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Webserver is a reusable web server component that creates and exports a NIC, public IP, and VM.
type Webserver struct {
	pulumi.ResourceState

	PublicIP         *network.PublicIp
	NetworkInterface *network.NetworkInterface
	VM               *compute.VirtualMachine
}

type WebserverArgs struct {
	// A required username for the VM login.
	Username pulumi.StringInput

	// A required encrypted password for the VM password.
	Password pulumi.StringInput

	// An optional boot script that the VM will use.
	BootScript pulumi.StringInput

	// An optional VM size; if unspecified, Standard_A0 (micro) will be used.
	VMSize pulumi.StringInput

	// A required Resource Group in which to create the VM
	ResourceGroupName pulumi.StringInput

	// A required Subnet in which to deploy the VM
	SubnetID pulumi.StringInput
}

// NewWebserver allocates a new web server VM, NIC, and public IP address.
func NewWebserver(ctx *pulumi.Context, name string, args *WebserverArgs, opts ...pulumi.ResourceOption) (*Webserver, error) {
	webserver := &Webserver{}
	err := ctx.RegisterComponentResource("ws-ts-azure-comp:webserver:WebServer", name, webserver, opts...)
	if err != nil {
		return nil, err
	}

	webserver.PublicIP, err = network.NewPublicIp(ctx, name+"-ip", &network.PublicIpArgs{
		ResourceGroupName: args.ResourceGroupName,
		AllocationMethod:  pulumi.String("Dynamic"),
	}, pulumi.Parent(webserver))
	if err != nil {
		return nil, err
	}

	webserver.NetworkInterface, err = network.NewNetworkInterface(ctx, name+"-nic", &network.NetworkInterfaceArgs{
		ResourceGroupName: args.ResourceGroupName,
		IpConfigurations: network.NetworkInterfaceIpConfigurationArray{
			network.NetworkInterfaceIpConfigurationArgs{
				Name:                       pulumi.String("webserveripcfg"),
				SubnetId:                   args.SubnetID.ToStringOutput(),
				PrivateIpAddressAllocation: pulumi.String("Dynamic"),
				PublicIpAddressId:          webserver.PublicIP.ID(),
			},
		},
	}, pulumi.Parent(webserver))
	if err != nil {
		return nil, err
	}

	vmSize := args.VMSize
	if vmSize == nil {
		vmSize = pulumi.String("Standard_A0")
	}

	// Now create the VM, using the resource group and NIC allocated above.
	webserver.VM, err = compute.NewVirtualMachine(ctx, name+"-vm", &compute.VirtualMachineArgs{
		ResourceGroupName:            args.ResourceGroupName,
		NetworkInterfaceIds:          pulumi.StringArray{webserver.NetworkInterface.ID()},
		VmSize:                       vmSize,
		DeleteDataDisksOnTermination: pulumi.Bool(true),
		DeleteOsDiskOnTermination:    pulumi.Bool(true),
		OsProfile: compute.VirtualMachineOsProfileArgs{
			ComputerName:  pulumi.String("hostname"),
			AdminUsername: args.Username,
			AdminPassword: args.Password.ToStringOutput(),
			CustomData:    args.BootScript.ToStringOutput(),
		},
		OsProfileLinuxConfig: compute.VirtualMachineOsProfileLinuxConfigArgs{
			DisablePasswordAuthentication: pulumi.Bool(false),
		},
		StorageOsDisk: compute.VirtualMachineStorageOsDiskArgs{
			CreateOption: pulumi.String("FromImage"),
			Name:         pulumi.String(fmt.Sprintf("%d", rangeIn(10000000, 99999999))),
		},
		StorageImageReference: compute.VirtualMachineStorageImageReferenceArgs{
			Publisher: pulumi.String("canonical"),
			Offer:     pulumi.String("UbuntuServer"),
			Sku:       pulumi.String("16.04-LTS"),
			Version:   pulumi.String("latest"),
		},
	}, pulumi.Parent(webserver), pulumi.DependsOn([]pulumi.Resource{webserver.NetworkInterface, webserver.PublicIP}))
	if err != nil {
		return nil, err
	}

	return webserver, nil
}

func (ws *Webserver) GetIPAddress(ctx *pulumi.Context) pulumi.StringOutput {
	// The public IP address is not allocated until the VM is running, so wait for that resource to create, and then
	// lookup the IP address again to report its public IP.
	ready := pulumi.All(ws.VM.ID(), ws.PublicIP.Name, ws.PublicIP.ResourceGroupName)
	return ready.ApplyT(func(args []interface{}) (string, error) {
		name := args[1].(string)
		resourceGroupName := args[2].(string)
		ip, err := network.GetPublicIP(ctx, &network.GetPublicIPArgs{
			Name:              name,
			ResourceGroupName: resourceGroupName,
		})
		if err != nil {
			return "", err
		}
		return ip.IpAddress, nil
	}).(pulumi.StringOutput)
}

func rangeIn(low, hi int) int {
	rand.Seed(time.Now().UnixNano())
	return low + rand.Intn(hi-low)
}

func ProvisionInfrastructure() error {

	ctx := context.Background()
	pulumiProjectName := StackInstance.Id
	pulumiStackName := "dev"

	// Specify a local backend instead of using the service.
	homeDir,_ := os.UserHomeDir()
	pulumiWorkspaceDir := "file://" + homeDir + "/.pulumi"
	project := auto.Project(workspace.Project{
		Name:    tokens.PackageName(pulumiProjectName),
		Runtime: workspace.NewProjectRuntimeInfo("go", nil),
		Backend: &workspace.ProjectBackend{
			URL: pulumiWorkspaceDir,
		},
	})

	// Setup a passphrase secrets provider and use an environment variable to pass in the passphrase.
	// secretsProvider := auto.SecretsProvider("passphrase")
	// envvars := auto.EnvVars(map[string]string{
	// 	// In a real program, you would feed in the password securely or via the actual environment.
	// 	"PULUMI_CONFIG_PASSPHRASE": "password",
	// })

	// stackSettings := auto.Stacks(map[string]workspace.ProjectStack{
	// 	"stackName" : {SecretsProvider: "passphrase"},
	// })

	// create or select a stack matching the specified name and project.
	// this will set up a workspace with everything necessary to run our inline program (deployFunc)
	stack, err := auto.UpsertStackInlineSource(ctx, pulumiStackName, pulumiProjectName, nil, project)
	if err != nil {
		fmt.Print("Stack creation error: ")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("finished creating stack ")

	w := stack.Workspace()
	if w == nil {
		fmt.Println("workspace is nil")
		os.Exit(1)
	}
	err = w.InstallPlugin(ctx, "azure", "v4.0.0")
	if err != nil {
		fmt.Printf("Failed to install program plugins: %v\n", err)
		os.Exit(1)
	}

	err = stack.SetConfig(ctx, "azure:location", auto.ConfigValue{Value: "westus"})
	if err != nil {
		fmt.Printf("Failed to set config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("ensuring network is configured...")
	subnetID, rgName, _ := EnsureNetwork(ctx, pulumiProjectName)

	// set out program for the deployment with the resulting network info
	w.SetProgram(GetDeployVMFunc(subnetID, rgName))

	fmt.Println("deploying vm webserver...")

	// wire up our update to stream progress to stdout
	stdoutStreamer := optup.ProgressStreams(os.Stdout)
	
	res, err := stack.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to deploy vm stack: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deployed server running at public IP %s\n", res.Outputs["ip"].Value.(string))
	return nil
}

// func rangeIn(low, hi int) int {
// 	rand.Seed(time.Now().UnixNano())
// 	return low + rand.Intn(hi-low)
// }

func GetDeployVMFunc(subnetID, rgName string) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		username := "pulumi"
		password, err := random.NewRandomPassword(ctx, "password", &random.RandomPasswordArgs{
			Length:  pulumi.Int(16),
			Special: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		server, err := NewWebserver(ctx, "vm-webserver", &WebserverArgs{
			Username: pulumi.String(username),
			Password: password.Result,
			BootScript: pulumi.String(fmt.Sprintf(`#!/bin/bash
	echo "Hello, from VMGR!" > index.html
	nohup python -m SimpleHTTPServer 80 &`)),
			ResourceGroupName: pulumi.String(rgName),
			SubnetID:          pulumi.String(subnetID),
		})
		if err != nil {
			return err
		}

		ctx.Export("ip", server.GetIPAddress(ctx))
		return nil
	}
}

// EnsureNetwork deploys the network stack if none exists, or simply returns the associated
// subnetID and resourceGroupName
func EnsureNetwork(ctx context.Context, projectName string) (string, string, error) {
	pulumiStackName := "networking"
	// create or select a stack with the inline networking program
	s, err := auto.UpsertStackInlineSource(ctx, pulumiStackName, projectName, DeployNetworkFunc)
	if err != nil {
		fmt.Printf("failed to create or select stack: %v\n", err)
	}

	outs, err := s.Outputs(ctx)
	if err != nil {
		fmt.Printf("failed to get networking stack outputs: %v\n", err)
		os.Exit(1)
	}
	rgName, rgOk := outs["rgName"].Value.(string)
	subnetID, sidOk := outs["subnetID"].Value.(string)
	if rgOk && sidOk && rgName != "" && subnetID != "" {
		return subnetID, rgName, nil
	}

	err = s.SetConfig(ctx, "azure:location", auto.ConfigValue{Value: "westus"})
	if err != nil {
		fmt.Printf("Failed to set config: %v\n", err)
		os.Exit(1)
	}

	// wire up our update to stream progress to stdout
	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	res, err := s.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to deploy network stack: %v\n", err)
		os.Exit(1)
	}
	return res.Outputs["subnetID"].Value.(string), res.Outputs["rgName"].Value.(string), nil
}

// DeployNetworkFunc is a pulumi program that sets up an RG, and virtual network.
func DeployNetworkFunc(ctx *pulumi.Context) error {
	rg, err := core.NewResourceGroup(ctx, "server-rg", nil)
	if err != nil {
		return err
	}

	network, err := network.NewVirtualNetwork(ctx, "server-network", &network.VirtualNetworkArgs{
		ResourceGroupName: rg.Name,
		AddressSpaces:     pulumi.StringArray{pulumi.String("10.0.0.0/16")},
		Subnets: network.VirtualNetworkSubnetArray{
			network.VirtualNetworkSubnetArgs{
				Name:          pulumi.String("default"),
				AddressPrefix: pulumi.String("10.0.1.0/24"),
			},
		},
	})

	subnetID := network.Subnets.Index(pulumi.Int(0)).Id().ApplyT(func(val *string) (string, error) {
		if val == nil {
			return "", nil
		}
		return *val, nil
	}).(pulumi.StringOutput)
	ctx.Export("subnetID", subnetID)
	ctx.Export("rgName", rg.Name)
	return nil
}
