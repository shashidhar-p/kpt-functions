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

var _ fn.Runner = &PodFunction{}

type PodFunction struct {
	data map[string]string
}

// Run is the main function logic.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r attributes
// `results` is the "ResourceList.Results" that you can write result info to.
func (r *PodFunction) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	nadConfig:= ""
	hasChanged := false
	for _, kubeObject := range items {
		if kubeObject.IsGVK("k8s.cni.cncf.io","v1","NetworkAttachmentDefinition"){
			nadConfig=kubeObject.GetName()
		}
	}
	if nadConfig == ""{
		*results = append(*results, fn.GeneralResult("NetworkAttachmentDefinition not found", fn.Error))
		return false
	}
    for _, kubeObject := range items {
		// Check for kind: ConfigMap
        if kubeObject.IsGVK("", "v1", "ConfigMap") {
			// Fetch Data field from configMap	
			data,_,_:= kubeObject.NestedStringMap("data")
			//Check image field exists
			_, ok := data["image"]
			if ok {
				// Set the kubeobject
				kubeObject.SetAPIVersion("v1")
				kubeObject.SetKind("Pod")
				kubeObject.SetAnnotation("k8s.v1.cni.cncf.io/networks",nadConfig)
				kubeObject.SetName(data["podName"])
				arrMaps:= []map[string]string{
					{"name":"test-container","image":data["image"]},
				}
				kubeObject.SetNestedField(arrMaps,"spec","containers")
				kubeObject.RemoveNestedField("data")
				hasChanged=true
			}
		}
    }
	if hasChanged {
		*results = append(*results, fn.GeneralResult("Created POD from configMap", fn.Info))
		return true
	}else {
		*results = append(*results, fn.GeneralResult("Failed to create POD", fn.Error))
		return false
	}
}

func main() {
	runner := fn.WithContext(context.Background(), &PodFunction{})
	if err := fn.AsMain(runner); err != nil {
		os.Exit(1)
	}
}
