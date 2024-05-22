#!/bin/bash

set -eo pipefail
set -x

./solrdump run -a $COMMAND -r kubedb-proxy-s3 -d solr-combined -n demo -l s3:/