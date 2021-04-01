 Test Ethclient - Contracts

## Contracts

The first step is to clone the github repository where the contracts are located:

`git clone https://github.com/hermeznetwork/contracts.git`

While the prepared deployment is not found to master, branch in repository must be changed:

`git checkout feature/newDeploymentScript-eth-mik` (tested with commit `0eccd237dda102a4ad788a6e11f98361d39d0d9c`)

Now, install the dependencies:

```
cd contracts/
yarn install
```

Go to where the deployment scripts for the test are found:

`cd scripts/ethclient-deployment/`

Now, in a terminal start a local blockchain with ganache:
```
../../node_modules/.bin/ganache-cli -d -m "explain tackle mirror kit van hammer degree position ginger unfair soup bonus" -p 8545 -l 12500000 -a 20 -e 10000 --allowUnlimitedContractSize --chainId 31337
```
Once ganache is ready, in another terminal run the deployment in the local ganache network:
```
npx buidler run --network localhostMnemonic test-deployment.js
```

An output file necessary for the next step is obtained: `deploy-output`.

> The files that appear in `hermez-node/eth/contracts` must be generated from the same contract that we deploy in this step

## Ethclient Test

Different environment variables are necessary to run this test.
They must be taken from the output file of the previous step.

They can be provided by file called `.env`:

```
GENESIS_BLOCK=98
AUCTION="0x317113D2593e3efF1FfAE0ba2fF7A61861Df7ae5"
AUCTION_TEST="0x2b7dEe2CF60484325716A1c6A193519c8c3b19F3"
TOKENHEZ="0x5D94e3e7aeC542aB0F9129B9a7BAdeb5B3Ca0f77"
HERMEZ="0x8EEaea23686c319133a7cC110b840d1591d9AeE0"
WDELAYER="0x5E0816F0f8bC560cB2B9e9C87187BeCac8c2021F"
WDELAYER_TEST="0xc8F466fFeF9E9788fb363c2F4fBDdF2cAe477805"
```

> An example is found in `hermez-node/eth/.env.example`

And then run test from `hermez-node/eth/`:

`INTEGRATION=1 go test`

Or they can be provided as a parameter in the command that runs the test:

`INTEGRATION=1 GENESIS_BLOCK=98 AUCTION="0x317113D2593e3efF1FfAE0ba2fF7A61861Df7ae5" AUCTION_TEST="0x2b7dEe2CF60484325716A1c6A193519c8c3b19F3" TOKENHEZ="0x5D94e3e7aeC542aB0F9129B9a7BAdeb5B3Ca0f77" HERMEZ="0x8EEaea23686c319133a7cC110b840d1591d9AeE0" WDELAYER="0x5E0816F0f8bC560cB2B9e9C87187BeCac8c2021F" WDELAYER_TEST="0xc8F466fFeF9E9788fb363c2F4fBDdF2cAe477805" go test`
