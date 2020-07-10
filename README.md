# Boatswain

Boatswain is a PoC implementation for doing EKS node maintenance/upgrades by
replacing nodes which were created from outdated launch templates.


## Commands

Boatswain provides three commands
* `check-ami` fetches the latest EKS AMI for a given version and displays ASGs
  which do not use the latest AMI
* `list-upgradable` lists nodes which are running off an outdated launch
  template version
* `upgrade` upgrades all nodes which are running off an outdated launch
  template version

## Configuration

Boatswain is configured via environment variables and command line flags

### Upgrade command

* `--single-node=NODE` allows replacing a single node, ignoring whether the
  node is using an outdated launch template.
* `--force-upgrade` replaces all nodes,  ignoring whether the nodes are using
  an outdated launch template. This option overrides `--single-node`, if both
  are set.


### AWS API client configuration

* `AWS_REGION` to set the AWS region in which to operate
* `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` as credentials
* `AWS_ASSUME_ROLE_ARN` to assume another role. This needs the full string,
  with form `arn:aws:iam::XXXXXXXXXXXX:role/OrganizationAccountAccessRole`
* the role to assume can also be provided via command line flag
  `--aws-assume-role-arn`.

### K8s API client configuration

We use the same K8s client-go mechanism that kubectl uses to find a kubeconfig
file. Environment variable `KUBECONFIG` wins, and otherwise the current
context in `~/.kube/config` is used.
