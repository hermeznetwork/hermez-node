#!/bin/sh

set -e

gen() {
    local pkg=$1
    local package=$2
    local folder=$3
    if [ -z "$3" ]; then
        folder=$package
    fi

    jq .abi "abi/${pkg}.json" > /tmp/${package}.abi
    abigen --abi /tmp/${package}.abi --pkg=${package} --out=${folder}/${package}.go
}

gen Hermez hermez
gen HermezAuctionProtocol auction
gen WithdrawalDelayer withdrawaldelayer
gen HEZ tokenhez
