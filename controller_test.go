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
	}{
		{
			"label is matched",
			"cloud.google.com/gke-nodepool",
			"default-pool",
			"us-central1-a",
			"my-node-name",
			"cloud.google.com/gke-nodepool=default-pool",
			"securityPolicyURL",
			"my-project-id",
			1,
		},
		{
			"label is not matched",
			"cloud.google.com/gke-nodepool",
			"default-pool",
			"us-central1-a",
			"my-node-name",
			"cloud.google.com/gke-nodepool=not-default-pool",
			"securityPolicyURL",
			"my-project-id",
			0,
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
			}

			reconcileReq := reconcile.Request{}

			if _, err := r.Reconcile(context.Background(), reconcileReq); err != nil {
				t.Errorf("error %v, want no error", err)
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
