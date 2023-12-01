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
	"encoding/json"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	
)

var _ fn.Runner = &NadFunction{}

type NadFunction struct {
	data map[string]string
}

// Run is the main function logic.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r attributes
// `results` is the "ResourceList.Results" that you can write result info to.
func (r *NadFunction) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	hasChanged := false
    for _, kubeObject := range items {
		// Check for kind: ConfigMap
        if kubeObject.IsGVK("", "v1", "ConfigMap") {
			var configContent map[string]interface{}
			// Fetch Data field from configMap	
			data,_,_:= kubeObject.NestedStringMap("data")
			config, ok := data["config"]	
			if !ok {
				continue
			}		
			err := json.Unmarshal([]byte(config), &configContent)
			if err != nil {
				return false			
			}
			// Check for NAD type: macvlan
			if(configContent["type"] == "macvlan"){
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
	if hasChanged {
		*results = append(*results, fn.GeneralResult("Created NetworkAttachmentDefinition from configMap", fn.Info))
		return true
	}else {
		*results = append(*results, fn.GeneralResult("Failed to create NetworkAttachmentDefinition", fn.Error))
		return false
	}
}

func main() {
	runner := fn.WithContext(context.Background(), &NadFunction{})
	if err := fn.AsMain(runner); err != nil {
		os.Exit(1)
	}
}
