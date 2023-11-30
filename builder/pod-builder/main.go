// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ fn.Runner = &YourFunction{}

// TODO: Change to your functionConfig "Kind" name.
type YourFunction struct {
	FnConfigBool bool
	FnConfigInt  int
	FnConfigFoo  string
}

type Pod struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind string `yaml:"kind"`
	ObjectMeta yaml.ObjectMeta `yaml:"metadata"`
	Spec struct {
		Containers []struct{
			Name string `yaml:"name"`
			Image string `yaml:"image"`
		} `yaml:"containers"`
	}
}

// Run is the main function logic.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r attributes
// `results` is the "ResourceList.Results" that you can write result info to.
func (r *YourFunction) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	// TODO: Write your code.
	for _, kubeObject := range items {
        if kubeObject.IsGVK("", "v1", "ConfigMap") {
			data,_,_:= kubeObject.NestedStringMap("data")
			pod := Pod{
				ApiVersion: "v1",
				Kind:       "Pod",
				ObjectMeta: yaml.ObjectMeta{
					NameMeta: yaml.NameMeta{
						Name: data["podName"],
					},
					Annotations:map[string]string{
						"k8s.v1.cni.cncf.io/networks": data["netAttachName"],
					},
				},
				Spec: struct {
					Containers []struct {
						Name  string `yaml:"name"`
						Image string `yaml:"image"`
					} `yaml:"containers"`
				}{
					Containers: []struct {
						Name  string `yaml:"name"`
						Image string `yaml:"image"`
					}{
						{
							Name:  data["containerName"],
							Image: data["image"],
						},
					},
				},
			}

			file, err := os.Create("shashi-pod.yaml")
			if err != nil {
				fmt.Printf("error creating YAML file: %v\n", err)
				return false
			}
			defer file.Close()
		
			err = yaml.NewEncoder(file).Encode(pod)
			if err != nil {
				fmt.Printf("error encoding YAML: %v\n", err)
				return false
			}
		
		}
	}


	return true
}

func main() {
	runner := fn.WithContext(context.Background(), &YourFunction{})
	if err := fn.AsMain(runner); err != nil {
		os.Exit(1)
	}
}