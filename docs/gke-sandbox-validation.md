# GKE sandbox validation

The `csb-google-gke` `sandbox-3-node` plan provisions a zonal GKE Standard
cluster for short-lived validation. Bindings contain cluster metadata and a
`gcloud container clusters get-credentials` recipe; they do not contain a
kubeconfig, access token, or static Kubernetes credential.

## Broker permissions

The broker service account requires permissions that cover both provisioning
and deprovisioning:

- `roles/container.admin` to manage GKE clusters.
- `roles/compute.viewer` to read GKE-managed instance groups during Terraform
  refresh, provisioning, and destroy.
- IAM permissions to create and delete the dedicated node service account,
  manage its project roles, and grant the broker `iam.serviceAccounts.actAs`
  on that account.

## Storage options

The plan enables the managed GCE Persistent Disk CSI driver by default. This
provides dynamically provisioned block storage without creating a disk until a
PersistentVolumeClaim requests one.

GKE cost allocation is also enabled by default so namespace and workload usage
can be included in exported billing data when the project has detailed billing
export configured.

Set `enable_filestore_csi` to `true` at provisioning time to enable the managed
Filestore CSI driver and its NFS storage classes. Enabling the driver does not
create a Filestore instance, but a claim using a Filestore storage class can
create billable infrastructure. Storage classes are cluster capabilities;
PersistentVolumes appear only after a claim is provisioned.

Example:

```sh
cf create-service csb-google-gke sandbox-3-node gke-sandbox \
  -c '{"cluster_name":"gke-sandbox","zone":"us-central1-a","enable_filestore_csi":true}'
```

## Status checks

Use filtered outputs when inspecting bindings. Do not print the complete
environment of an application that has other service bindings.

```sh
cf service gke-sandbox
gcloud container clusters describe gke-sandbox \
  --project "$GOOGLE_PROJECT" \
  --zone us-central1-a \
  --format='yaml(name,status,currentNodeCount)'
gcloud container clusters get-credentials gke-sandbox \
  --project "$GOOGLE_PROJECT" \
  --zone us-central1-a
kubectl get nodes
kubectl get storageclass
kubectl get persistentvolume
```

## Acceptance criteria

- Cloud Foundry reports `create succeeded`.
- GKE reports the cluster as `RUNNING` with `currentNodeCount: 3`.
- GKE reports cost allocation as enabled.
- `kubectl` authenticates using ambient Google Cloud credentials.
- Exactly three nodes report the `Ready` condition as `True`.
- The binding's `normalized_binding_json` reports provider `gcp`, connection
  type `cluster`, the expected project, zone, cluster name, and the ambient-auth
  command; it contains no token, key, or kubeconfig.
- With the default plan, at least one StorageClass uses provisioner
  `pd.csi.storage.gke.io`.
- When `enable_filestore_csi` is `true`, a Filestore StorageClass using
  provisioner `filestore.csi.storage.gke.io` is available.
- A test PersistentVolumeClaim reaches `Bound`, its pod can write and read a
  marker, and deleting the claim removes the dynamically provisioned volume
  according to its reclaim policy.
- Cloud Foundry reports successful deletion, and subsequent GKE cluster, node
  VM, and dedicated node-service-account lookups return no resources.

The PVC workload check is required before claiming storage support as live
validated. An empty PersistentVolumes table before any claim is expected.
