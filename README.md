## Overview
The Cloud Armor Node Controller can help provide UDP DDoS protection for game customers trying to adopt GCP. It allows users to apply security policies filtered by node labels. The Cloud Armor Node Controller acts as a temporary solution until GKE provides integration of DDoS protection for GKE Node IPs.

## Deployment
Steps to run the Cloud Armor Node Controller:

1. Build and push the Cloud Armor Node Controller image to the repository of your choice in Artifact Registry:
    ```shell
    export REPOSITORY=<REPOSITORY>
    docker build --tag=$REPOSITORY/cloud-armor-node-controller:0.1
    docker push $REPOSITORY/cloud-armor-node-controller:0.1
    ```

    Or alternatively with Cloud Build:
    ```shell
    export REPOSITORY=<REPOSITORY>
    make cloud-build REPOSITORY=$REPOSITORY
    ```

2. Give IAM roles `roles/compute.instanceAdmin.v1` and `roles/compute.networkAdmin` to Kubernetes Service Account `cloud-armor-node-controller-sa`:
    ```shell
    export PROJECT_ID=<PROJECT_ID>
    export PROJECT_NUMBER=<PROJECT_NUMBER>
    export KUBERNETES_SERVICE_ACCOUNT=principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/cloud-armor-node-controller/sa/cloud-armor-node-controller-sa

    gcloud projects add-iam-policy-binding projects/$PROJECT_ID --role=roles/compute.instanceAdmin.v1 --member=$KUBERNETES_SERVICE_ACCOUNT --condition=None
    gcloud projects add-iam-policy-binding projects/$PROJECT_ID --role=roles/compute.networkAdmin --member=$KUBERNETES_SERVICE_ACCOUNT --condition=None
    ```

3. To deploy, edit the `deployment.yaml` file: 
    * Replace `YOUR_REPOSITORY_HERE` with the repository chosen above.
    * Replace `SELECTOR` to be a Kubernetes label selector, for example - `cloud.google.com/gke-nodepool=default-pool`.
    * Replace `SECURITY_POLICY` to be the name of a security policy you created(Note: just the name, not the full path). 
    
    Then, deploy the Cloud Armor Node Controller in your cluster in `cloud-armor-node-controller` namespace with:
    ```shell
    kubectl apply -f deployment.yaml
    ```

## Notes

* The Cloud Armor Node Controller will set the security policy of the nodes that are selected by the `SELECTOR` enironment variable exactly once. More specifically, it will set the security policy the first time it sees a node create/update event. If later the security policy on the node is changed by some other means, such as with `gcloud` commands, then the Cloud Armor Node Controller will not overwrite the security policy immediately; instead, it will overwrite the security policy if the Cloud Armor Node Controller is restarted, such as when a pod is rescheduled.

* To assign different security policies with different label selectors, create multiple deployments of Cloud Armor Node Controllers, each with its own label selector and security policy.

* If a node is selected by multiple label selectors at the same time and thus eligible to be applied with multiple security policies, any one of the security policies may be set on the node.