# Boatswain

Boatswain is a PoC implementation for doing EKS node maintenance/upgrades by
replacing nodes which were created from outdated launch templates.


## Configuration

Boatswain is configured via environment variables

The variable `NODE` can be set to a node name to only operate on that node.

### AWS API client configuration

* `AWS_REGION` to set the AWS region in which to operate
* `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` as credentials
* `AWS_ASSUME_ROLE_ARN` to assume another role. This needs the full string,
  with form `arn:aws:iam::XXXXXXXXXXXX:role/OrganizationAccountAccessRole`

### K8s API client configuration

We use the same K8s client-go mechanism that kubectl uses to find a kubeconfig
file. Environment variable `KUBECONFIG` wins, and otherwise the current
context in `~/.kube/config` is used.
