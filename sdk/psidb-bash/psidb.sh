#!/usr/bin/env bash

psidb_create_node() {
  p=$1
  shift

  http --verify=no POST "$PSIDB_ENDPOINT/v1/psi/$p" Accept:application/json Content-Type:application/json "$@"
}

psidb_rpc_node() {
  p=$1
  iface=$2
  act=$3
  args=$(echo -n "${4:-"{}"}" | base64)
  shift 3

  http --verbose --verify=no POST "$PSIDB_ENDPOINT/rpc/v1" jsonrpc=2.0 id=1 method=NodeService.CallNodeAction \
    "params[path]=$p" "params[interface]=$iface" "params[action]=$act" "params[args]=$args"
}
