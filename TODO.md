# TODO

* Better error handling
* Prometheus metrics / alerts when upgrade fails
* Improve node replacement logic
  -> Use detachinstances followed by drain&terminate
* Consider comparing node launch template AMI with latest launch template AMI
  instead of just comparing launch template versions to avoid node
  replacements when only userdata changes
  -> use aws ssm get-parameter "/aws/service/eks/optimized-ami/${EKS_VERSION}/amazon-linux-2/recommended/image_id"
* Find a way to trigger upgrades at a given time

* More resilient drain
