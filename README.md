## Overview
The Cloud Armor Node Controller can help provide UDP DDoS protection for game customers trying to adopt GCP. It allows users to apply security policies filtered by node labels. The Cloud Armor Node Controller acts as a temporary solution until GKE provides integration of DDoS protection for GKE Node IPs.

## Deployment
Steps to run the Cloud Armor Node Controller:

1. Build the Cloud Armor Node Controller by running `docker build --tag=$REPOSITORY/cloud-armor-node-controller:0.1`.

2. Push the docker image to your desired Artifact Registry repository by running `docker push $REPOSITORY/cloud-armor-node-controller:0.1`.

3. Give IAM roles `roles/compute.instanceAdmin.v1` and `roles/compute.networkAdmin` to Kubernetes Service Account `cloud-armor-node-controller-sa` (The full name of the KSA should be in the format of `principal://iam.googleapis.com/projects/PROJECT_ID/locations/global/workloadIdentityPools/PROJECT_ID.svc.id.goog/subject/ns/cloud-armor-node-controller/sa/cloud-armor-node-controller-sa`).

3. Deploy the Cloud Armor Node Controller in your cluster in `cloud-armor-node-controller` namespace by running `kubectl apply -f deployment.yaml`.
