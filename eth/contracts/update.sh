#!/bin/sh

set -e

gen() {
    local name=$1
    local pkg=$2
    local package=$3
    local folder=$4
    if [ -z "$4" ]; then
        folder=$package
    fi

    jq .abi "${CONTRACTS}/artifacts/contracts/${name}/${pkg}.sol/${pkg}.json" > /tmp/${name}.abi
    abigen --abi /tmp/${name}.abi --pkg=${package} --out=${folder}/${package}.go
}

if [ "$1" = "" ]; then
    echo "Usage: $0 CONTRACTS_REPO_PATH"
    exit 1
fi

CONTRACTS="$1"

gen hermez Hermez hermez
gen auction HermezAuctionProtocol auction
gen withdrawalDelayer WithdrawalDelayer withdrawaldelayer
gen HEZ HEZ tokenhez
