# [WIP] RedPanda Operator

## How-to start:
```make test```

or manual testing:
1. Get running cluster (for example using `kind`)
2. Make sure you `kubectl` is using correct cluster context.
3. Install CRDs and take care of permissions: `make generate && make manifests && make install`
4. `make run` - this will run operator outside k8s cluster for easy troubleshooting.
5. `kubectl apply -f testdata/redpanda-cluster.yaml`

## TODO: 
- [x] basic operator using `StatefulSet` as backend resource.

     - [ ] improve testing coverage: see `controllers/*_test.go` and [kubebuilder docs](https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html)
     - [ ] finalizers logic
     - [ ] defaulters logic
     - [ ] validations
     - [ ] scale-out support
     - [ ] refactor configuration into templated `ConfigMap` for operator
     - [ ] consider whether using `StatefulSets` does not limit us:
     
        Context: [Banzai Cloud - Kafka Operator: To StatefulSet or not to StatefulSet: and why it matters](https://banzaicloud.com/blog/kafka-operator/#to-statefulset-or-not-to-statefulset-and-why-it-matters)
        ```
       With StatefulSet we get:
       
       - unique Broker IDs generated during Pod startup
       - networking between brokers with headless services
       - unique Persistent Volumes for Brokers
       
       Using StatefulSet we lose:
       
       - the ability to modify the configuration of unique Brokers
       - to remove a specific Broker from a cluster (StatefulSet always removes the most recently created Broker)
       - to use multiple, different Persistent Volumes for each Broker
       ```

- [ ] other considerations:
    - [ ] design operator to be open for multiple backends (AWS K8S Controller, Terraform (?))
   
- [ ] add support for monitoring with Prometheus Operator