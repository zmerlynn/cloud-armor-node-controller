package main

import (
	"context"
	"fmt"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type runtimeClientImpt struct {
	labels map[string]string
}

func (r *runtimeClientImpt) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	fmt.Println("Called Get")
	obj.SetLabels(r.labels)
	return nil
}

type instanceClientImpl struct {
	calledWithSecurityPolicy string
}

func (i *instanceClientImpl) SetSecurityPolicy(ctx context.Context, req *computepb.SetSecurityPolicyInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error) {
	fmt.Println("Called SetSecurityPolicy")
	i.calledWithSecurityPolicy = *req.InstancesSetSecurityPolicyRequestResource.SecurityPolicy
	return nil, nil
}

func TestReconcile(t *testing.T) {
	labels := make(map[string]string)
	labels["cloud.google.com/gke-nodepool"] = "default-pool"
	rc := runtimeClientImpt{labels: labels}
	ic := instanceClientImpl{}

	s := "cloud.google.com/gke-nodepool=default-pool"
	spURL := "securityPolicy"
	r := &reconcileNode{
		client:            &rc,
		instanceClient:    &ic,
		selector:          &s,
		securityPolicyURL: spURL,
	}

	reconcileReq := reconcile.Request{}

	_, err := r.Reconcile(context.Background(), reconcileReq)
	if err != nil {
		t.Error(err)
	}

	if ic.calledWithSecurityPolicy != spURL {
		t.Errorf("got %q, wanted %q", ic.calledWithSecurityPolicy, spURL)
	}
}
