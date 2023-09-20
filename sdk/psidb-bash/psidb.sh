#!/usr/bin/env bash

psidb_create_node() {
  p=$1
  shift

  http --verify=no POST "$PSIDB_ENDPOINT/v1/psi/$p" Accept:application/json Content-Type:application/json "$@"
}

psidb_rpc_node() {
  local p iface act params

  p=$1
  iface=$2
  act=$3
  shift 3;

  params=()

  for arg in "${@:-}"; do
    key="${arg%=*}"
    value="${arg#*=}"
    params+=("params[args][$key]=$value")
  done

  http --verbose --verify=no POST "$PSIDB_ENDPOINT/rpc/v1" jsonrpc=2.0 id=1 method=NodeService.CallNodeAction \
    "params[path]=$p" "params[interface]=$iface" "params[action]=$act" "${params[@]}"
}
