#!/bin/sh

set -e

gen() {
    local name=$1
    local pkg=$2
    local folder=$3
    if [ -z "$3" ]; then
        folder=$name
    fi

    jq .abi "${CONTRACTS}/artifacts/contracts/${name}/${pkg}.sol/${pkg}.json" > /tmp/${name}.abi
    abigen --abi /tmp/${name}.abi --pkg=${pkg} --out=${folder}/${pkg}.go
}

if [ "$1" = "" ]; then
    echo "Usage: $0 CONTRACTS_REPO_PATH"
    exit 1
fi

CONTRACTS="$1"

gen hermez Hermez
gen auction HermezAuctionProtocol
gen withdrawalDelayer WithdrawalDelayer
gen HEZ HEZ tokenHEZ
