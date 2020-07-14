# TODO

Notes on the shortcomings of the implementation. These should be translated
into issues at some point.

* [ ] Better error handling
* [ ] Prometheus metrics / alerts when replacing node fails
* [ ] Consider comparing node launch template AMI with latest launch template AMI instead of just comparing launch template versions to avoid node replacements when only userdata changes. This was discussed, but discarded, as the current approach also allows replacing nodes to have all nodes pick up new startup configurations even if the EKS AMI does not change.
* [ ] Find a way to trigger upgrades at a given time. Not necessarily in scope for Boatswain
* [ ] More resilient drain. This may be fixed in the latest version
* [ ] Maybe disable cluster-autoscaler during upgrade -- observe how current implementation behaves first
* [ ] Figure out how to handle spot instance ASGs -- they don't have launch templates but launch configurations
