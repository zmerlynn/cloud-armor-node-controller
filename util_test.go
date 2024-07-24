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
	"testing"
)

func TestMakeSecurityPolicyURL(t *testing.T) {
	var tests = []struct {
		name           string
		projectID      string
		region         string
		securityPolicy string
		expectedURL    string
	}{
		{
			name:           "non-empty securitypolicy",
			projectID:      "my-project-id",
			region:         "us-central1",
			securityPolicy: "my-security-policy",
			expectedURL:    "https://www.googleapis.com/compute/v1/projects/my-project-id/regions/us-central1/securityPolicies/my-security-policy",
		},
		{
			name:           "empty securitypolicy",
			projectID:      "my-project-id",
			region:         "us-central1",
			securityPolicy: "",
			expectedURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spURL := makeSecurityPolicyURL(tt.projectID, tt.region, tt.securityPolicy)
			if tt.expectedURL != spURL {
				t.Errorf("wrong security policy url: %s, want %s", spURL, tt.expectedURL)
			}
		})
	}
}

func TestExtractRegionFromZone(t *testing.T) {
	var tests = []struct {
		name           string
		zone           string
		expectedRegion string
	}{
		{
			name:           "extract region from zone",
			zone:           "us-central1-a",
			expectedRegion: "us-central1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := extractRegionFromZone(tt.zone)
			if tt.expectedRegion != region {
				t.Errorf("wrong region: %s, want %s", region, tt.expectedRegion)
			}
		})
	}
}
