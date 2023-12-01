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

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	
)

var _ fn.Runner = &YourFunction{}

// TODO: Change to your functionConfig "Kind" name.
type YourFunction struct {
	FnConfigBool bool
	FnConfigInt  int
	FnConfigFoo  string
	data map[string]string
}

// Run is the main function logic.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r attributes
// `results` is the "ResourceList.Results" that you can write result info to.
func (r *YourFunction) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	hasChanged := false
    for _, kubeObject := range items {
        if kubeObject.IsGVK("", "v1", "ConfigMap") {	
			data,_,_:= kubeObject.NestedStringMap("data")
			config:= data["config"]

			funConf,_,_:= functionConfig.NestedStringMap("data")
			if(kubeObject.GetName() == funConf["resourceName"]){
				// Set the kubeobject
				kubeObject.SetAPIVersion("k8s.cni.cncf.io/v1")
				kubeObject.SetKind("NetworkAttachmentDefinition")
				kubeObject.SetName(data["netAttachName"])
				kubeObject.SetNestedField(&config,"spec","config")
				kubeObject.RemoveNestedField("data")
				hasChanged=true
			}
        }
    }
	if(hasChanged){
		*results = append(*results, fn.GeneralResult("Created NetworkAttachmentDefinition from configMap", fn.Info))
		return true
	}else {
		*results = append(*results, fn.GeneralResult("No resource found with the given name", fn.Error))
		return false
	}
	
}

func main() {
	runner := fn.WithContext(context.Background(), &YourFunction{})
	if err := fn.AsMain(runner); err != nil {
		os.Exit(1)
	}
}
