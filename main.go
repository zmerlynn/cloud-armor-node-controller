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
	"flag"
	"fmt"
	"os"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/metadata"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func makeSecurityPolicyURL(projectID, region, securityPolicy string) string {
	if securityPolicy == "" {
		return ""
	}
	return "https://www.googleapis.com/compute/v1/projects/" + projectID + "/regions/" + region + "/securityPolicies/" + securityPolicy
}

func init() {
	log.SetLogger(zap.New())
}

func main() {
	entryLog := log.Log.WithName("entrypoint")
	// Parse Env Flags
	selector := flag.String("selector", "", "The node label selector")
	securityPolicy := flag.String("securityPolicy", "", "The URL of the security policy. Empty string by default, which will unset the security policy of the nodes selected")
	flag.Parse()

	if es := os.Getenv("SELECTOR"); es != "" {
		selector = &es
	}
	if esp := os.Getenv("SECURITY_POLICY"); esp != "" {
		securityPolicy = &esp
	}

	// Set up client to call GCE compute APIS.
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		entryLog.Error(err, "unable to run instances client")
		os.Exit(1)
	}
	defer instancesClient.Close()

	// Set up metadata client
	metadataClient := metadata.NewClient(nil)

	// Get project id, and region to make security policy url.
	projectID, err := metadataClient.ProjectID()
	if err != nil {
		entryLog.Error(err, "unable to run get project ID")
		os.Exit(1)
	}
	fmt.Println("project ID is ", projectID)
	zone, err := metadataClient.Zone()
	if err != nil {
		entryLog.Error(err, "unable to run get zone")
		os.Exit(1)
	}
	fmt.Println("zone is ", zone)

	parts := strings.Split(zone, "-")
	region := parts[0] + "-" + parts[1]
	fmt.Println("region is ", region)

	// Setup a Manager
	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// Setup a new controller to reconcile Nodes
	entryLog.Info("Setting up controller")
	c, err := controller.New("foo-controller", mgr, controller.Options{
		Reconciler: &reconcileNode{
			client:            mgr.GetClient(),
			instanceClient:    instancesClient,
			selector:          *selector,
			securityPolicyURL: makeSecurityPolicyURL(projectID, region, *securityPolicy),
			projectID:         projectID,
		},
	})
	if err != nil {
		entryLog.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}

	// Watch Nodes and enqueue Node object key
	if err := c.Watch(source.Kind(mgr.GetCache(), &corev1.Node{}, &handler.TypedEnqueueRequestForObject[*corev1.Node]{})); err != nil {
		entryLog.Error(err, "unable to watch Nodes")
		os.Exit(1)
	}

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
