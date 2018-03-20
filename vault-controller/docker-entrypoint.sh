#!/bin/bash

set -m

####################################################################################################
# This script preps the vault-client image
#
####################################################################################################

echo "Vault ready"
exec /controller &
pid=$!
trap 'kill -SIGTERM $pid' EXIT
fg %1
