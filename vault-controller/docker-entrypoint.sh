#!/bin/bash

set -m

####################################################################################################
# This script preps the vault-client image
#
####################################################################################################

# VAULT_TOKEN overrides VAULT_TOKEN_FILE
if [[ ! ${VAULT_TOKEN} =~ ^\{?[A-F0-9a-f]{8}-[A-F0-9a-f]{4}-[A-F0-9a-f]{4}-[A-F0-9a-f]{4}-[A-F0-9a-f]{12}\}?$ ]] && [[ ! -z ${VAULT_TOKEN_FILE} ]]; then
  echo "Using VAULT_TOKEN_FILE: ${VAULT_TOKEN_FILE}..."
  [ -f ${VAULT_TOKEN_FILE} ] && export VAULT_TOKEN=`cat ${VAULT_TOKEN_FILE}`
fi

# VAULT_UNSEAL_KEYS overrides VAULT_UNSEAL_KEYS_FILE
if [[ ! -z ${VAULT_UNSEAL_KEYS_FILE} && "x${VAULT_UNSEAL_KEYS}x" == "xx" ]]; then
    for f in ${VAULT_UNSEAL_KEYS_FILE}; do
        [ -f $f ] && export VAULT_UNSEAL_KEYS="`cat $f`,$VAULT_UNSEAL_KEYS"
    done
    unset f
fi

echo "Vault Controller starting..."
exec /controller "$@" &
if [ ${DEBUG} == "true" ]; then
    set
fi
pid=$!
trap 'kill -SIGTERM $pid' EXIT
fg %1
