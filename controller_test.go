package main

import (
	"context"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
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
	calledWithProjectID        string
	calledWithZone             string
	calledWithInstanceName     string
	calledWithSecurityPolicy   string
	calledWithNetworkInterface []string
	timesCalled                int
}

func (i *instanceClientImpl) SetSecurityPolicy(ctx context.Context, req *computepb.SetSecurityPolicyInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error) {
	i.calledWithProjectID = req.Project
	i.calledWithZone = req.Zone
	i.calledWithInstanceName = req.Instance
	i.calledWithSecurityPolicy = *req.InstancesSetSecurityPolicyRequestResource.SecurityPolicy
	i.calledWithNetworkInterface = req.InstancesSetSecurityPolicyRequestResource.NetworkInterfaces
	i.timesCalled++
	return nil, nil
}

func TestReconcile(t *testing.T) {
	labels := make(map[string]string)
	labels["cloud.google.com/gke-nodepool"] = "default-pool"
	labels["topology.gke.io/zone"] = "us-central1-a"
	name := "my-node"
	rc := runtimeClientImpt{labels: labels, nodeName: name}
	ic := instanceClientImpl{}

	s := "cloud.google.com/gke-nodepool=default-pool"
	spURL := "securityPolicy"
	projectID := "my-project-id"
	r := &reconcileNode{
		client:            &rc,
		instanceClient:    &ic,
		selector:          s,
		securityPolicyURL: spURL,
		projectID:         projectID,
	}

	reconcileReq := reconcile.Request{}

	if _, err := r.Reconcile(context.Background(), reconcileReq); err != nil {
		t.Errorf("error %v, want no error", err)
	}

	if ic.calledWithSecurityPolicy != spURL {
		t.Errorf("wrong security policy: %s, want %s", ic.calledWithSecurityPolicy, spURL)
	}

	if ic.calledWithProjectID != projectID {
		t.Errorf("wrong project id: %s, want %s", ic.calledWithProjectID, projectID)
	}

	if ic.calledWithInstanceName != name {
		t.Errorf("wrong instance name: %s, want %s", ic.calledWithInstanceName, name)
	}

	if len(ic.calledWithNetworkInterface) != 1 || ic.calledWithNetworkInterface[0] != "nic0" {
		t.Errorf("wrong network interface: %v, want only 1 element of nic0", ic.calledWithNetworkInterface)
	}

	if ic.timesCalled != 1 {
		t.Errorf("wrong number of times called: %d, want %d", ic.timesCalled, 1)
	}
}
