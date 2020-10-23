 Test Ethclient - Contracts

## Contracts

The first step is to clone the github repository where the contracts are located:

`git clone https://github.com/hermeznetwork/contracts.git`

While the prepared deployment is not found to master, branch in repository must be changed:

`git checkout feature/ethclient-test-deployment`

Now, install the dependencies:

`npm i`

Go to where the deployment scripts for the test are found:

`cd scripts/ethclient-test-deployment`

Now, a bash script has to be run to do the deployment:
`./test-deployment`

This bash file follows these steps:
- `npx builder node`: a local blockchain to do our tests
- `npx buidler run --network localhost test-deployment.js`: run the deployment on the local blockchain


An output file necessary for the next step is obtained: `deploy-output`.

> The files that find in `/eth/contracts` must be obtained from the same contract that we deploy in this step
## Ethclient Test

Different environment variables are necessary to run this test.
They must be taken from the output file of the previous step.

They can be provided by file called `.env`:

```
GENESIS_BLOCK=97
AUCTION="0x038B86d9d8FAFdd0a02ebd1A476432877b0107C8"
AUCTION_TEST="0xEcc0a6dbC0bb4D51E4F84A315a9e5B0438cAD4f0"
TOKENHEZ="0xf4e77E5Da47AC3125140c470c71cBca77B5c638c"
HERMEZ="0xD6C850aeBFDC46D7F4c207e445cC0d6B0919BDBe"
WDELAYER="0x500D1d6A4c7D8Ae28240b47c8FCde034D827fD5e"
WDELAYER_TEST="0x1d80315fac6aBd3EfeEbE97dEc44461ba7556160"
```

> An example is found in `/etc/.env.example`

And then run test:

`INTEGRATION=1 go test`

Or they can be provided as a parameter in the command that runs the test:

`INTEGRATION=1 GENESIS_BLOCK=97 AUCTION="0x038B86d9d8FAFdd0a02ebd1A476432877b0107C8" AUCTION_TEST="0xEcc0a6dbC0bb4D51E4F84A315a9e5B0438cAD4f0" TOKENHEZ="0xf4e77E5Da47AC3125140c470c71cBca77B5c638c" HERMEZ="0xD6C850aeBFDC46D7F4c207e445cC0d6B0919BDBe" WDELAYER="0x500D1d6A4c7D8Ae28240b47c8FCde034D827fD5e" WDELAYER_TEST="0x1d80315fac6aBd3EfeEbE97dEc44461ba7556160" go test`
