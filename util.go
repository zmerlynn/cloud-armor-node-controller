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
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func makeSecurityPolicyURL(projectID, region, securityPolicy string) string {
	if securityPolicy == "" {
		return ""
	}
	return "https://www.googleapis.com/compute/v1/projects/" + projectID + "/regions/" + region + "/securityPolicies/" + securityPolicy
}

func extractRegionFromZone(zone string) string {
	parts := strings.Split(zone, "-")
	return parts[0] + "-" + parts[1]
}
