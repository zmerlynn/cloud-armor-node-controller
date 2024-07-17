/*
Copyright 2024 The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type runtimeClient interface {
	Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error
}

type instanceClient interface {
	SetSecurityPolicy(ctx context.Context, req *computepb.SetSecurityPolicyInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error)
}

type reconcileNode struct {
	// client can be used to retrieve objects from the APIServer.
	client            runtimeClient
	instanceClient    instanceClient
	selector          labels.Selector
	securityPolicyURL string
	projectID         string
}

// Assert reconcileNode implements reconcile.Reconciler.
var _ reconcile.Reconciler = &reconcileNode{}

func (r *reconcileNode) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Node from the cache
	node := &corev1.Node{}
	if err := r.client.Get(ctx, request.NamespacedName, node); err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch Node: %+v", err)
	}

	// Print the node
	name := node.Name
	zone := node.Labels["topology.gke.io/zone"]
	log.Info("Reconciling Node", "nodeName", name, "nodeZone", zone)

	// Check if node label matches the selector
	isMatch := r.selector.Matches(labels.Set(node.Labels))
	log.Info("Should apply security policy", "value", isMatch)

	if !isMatch {
		return reconcile.Result{}, nil
	}

	// Set Google Cloud Armor security policy for the instance
	networkInterfaces := []string{"nic0"}
	spReq := &computepb.SetSecurityPolicyInstanceRequest{
		Project:  r.projectID,
		Zone:     zone,
		Instance: name,
		InstancesSetSecurityPolicyRequestResource: &computepb.InstancesSetSecurityPolicyRequest{
			SecurityPolicy:    &r.securityPolicyURL,
			NetworkInterfaces: networkInterfaces,
		},
	}

	if _, err := r.instanceClient.SetSecurityPolicy(ctx, spReq); err != nil {
		return reconcile.Result{}, fmt.Errorf("could not set security policy of instance: %+v", err)
	}

	log.Info("Set security policy successfully", "node", name, "policy", r.securityPolicyURL)

	return reconcile.Result{}, nil
}
