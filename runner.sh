#!/bin/bash

set -eo pipefail
set -x

./solrdump run -r kubedb-linode-s3 -d solr-combined -n demo -l s3:/hello