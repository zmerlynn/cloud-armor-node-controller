// Copyright 2024 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/googleapis/gax-go/v2"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type runtimeClientImpt struct {
	labels   map[string]string
	nodeName string
}

func (r *runtimeClientImpt) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	obj.SetLabels(r.labels)
	obj.SetName(r.nodeName)
	return nil
}

type instanceClientImpl struct {
	request     *computepb.SetSecurityPolicyInstanceRequest
	timesCalled int
}

func (i *instanceClientImpl) SetSecurityPolicy(ctx context.Context, req *computepb.SetSecurityPolicyInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error) {
	i.request = req
	i.timesCalled++
	return nil, nil
}

func TestReconcileTable(t *testing.T) {
	var tests = []struct {
		name              string
		label             string
		value             string
		zone              string
		nodeName          string
		selector          string
		securityPolicyURL string
		projectID         string
		timesCalled       int
		timesToReconcile  int
	}{
		{
			name:              "label is matched",
			label:             "cloud.google.com/gke-nodepool",
			value:             "default-pool",
			zone:              "us-central1-a",
			nodeName:          "my-node-name",
			selector:          "cloud.google.com/gke-nodepool=default-pool",
			securityPolicyURL: "securityPolicyURL",
			projectID:         "my-project-id",
			timesCalled:       1,
			timesToReconcile:  1,
		},
		{
			name:              "label is not matched",
			label:             "cloud.google.com/gke-nodepool",
			value:             "default-pool",
			zone:              "us-central1-a",
			nodeName:          "my-node-name",
			selector:          "cloud.google.com/gke-nodepool=not-default-pool",
			securityPolicyURL: "securityPolicyURL",
			projectID:         "my-project-id",
			timesCalled:       0,
			timesToReconcile:  1,
		},
		{
			name:              "a node's security policy should only be set once",
			label:             "cloud.google.com/gke-nodepool",
			value:             "default-pool",
			zone:              "us-central1-a",
			nodeName:          "my-node-name",
			selector:          "cloud.google.com/gke-nodepool=default-pool",
			securityPolicyURL: "securityPolicyURL",
			projectID:         "my-project-id",
			timesCalled:       1,
			timesToReconcile:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeLabels := make(map[string]string)
			nodeLabels[tt.label] = tt.value
			nodeLabels["topology.gke.io/zone"] = tt.zone
			name := tt.nodeName
			parsedSelector, err := labels.Parse(tt.selector)
			if err != nil {
				t.Errorf("error parsing selector %v, want no error", err)
			}
			rc := runtimeClientImpt{labels: nodeLabels, nodeName: name}
			ic := instanceClientImpl{}

			r := &reconcileNode{
				client:            &rc,
				instanceClient:    &ic,
				selector:          parsedSelector,
				securityPolicyURL: tt.securityPolicyURL,
				projectID:         tt.projectID,
				processedNodes:    make(map[string]bool),
			}

			reconcileReq := reconcile.Request{}

			for i := 0; i < tt.timesToReconcile; i++ {
				if _, err := r.Reconcile(context.Background(), reconcileReq); err != nil {
					t.Errorf("error %v, want no error", err)
				}
			}

			if ic.timesCalled != tt.timesCalled {
				t.Errorf("wrong number of times called: %d, want %d", ic.timesCalled, tt.timesCalled)
			}

			if ic.timesCalled == 0 {
				// Skip checking the SetSecurityPolicyInstanceRequest because there is none.
				return
			}

			expectedRequest := computepb.SetSecurityPolicyInstanceRequest{
				Project:  tt.projectID,
				Zone:     tt.zone,
				Instance: tt.nodeName,
				InstancesSetSecurityPolicyRequestResource: &computepb.InstancesSetSecurityPolicyRequest{
					SecurityPolicy:    &tt.securityPolicyURL,
					NetworkInterfaces: []string{"nic0"},
				},
			}
			if diff := cmp.Diff(&expectedRequest, ic.request,
				cmpopts.IgnoreUnexported(computepb.SetSecurityPolicyInstanceRequest{},
					computepb.InstancesSetSecurityPolicyRequest{})); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
