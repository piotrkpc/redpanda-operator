#!/usr/bin/env bash

kubectl delete redpandaclusters.eventstream.vectorized.io redpanda-test && kubectl delete cm redpanda-testbase-config && kubectl delete svc redpanda-test && kubectl delete pvc redpanda-data-redpanda-test-{0..2}