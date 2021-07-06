# Setup Hermez Node

This tutorial will help you to setup an Hermez Node and run as sync mode.

## Golang

- [ ] [Install](https://golang.org/doc/install#install) Golang
- [ ] [Create a](https://golang.org/doc/gopath_code#GOPATH) GOPATH environment
- [ ] Set GOPATH/bin into PATH machine

    ```command
    export PATH=$PATH:$GOPATH/bin
    ```

## Go-Ethereum

- [ ] Clone [GETH](https://github.com/ethereum/go-ethereum) project and do a `make all` command into project folder.
  - Need have to gcc on PATH.

    On Ubuntu/Debian you need install `build-essential`

    ```command
    sudo apt-get update
    sudo apt-get install build-essential
    ```

- [ ] Add `build/bin` on PATH, after `make install` command inside of GETH
folder.

    ```command
    export GETH_PATH=$GOPATH/src/github.com/ethereum/go-ethereum/
    export PATH=$PATH:$GETH_PATH/build/bin
    ```

- [ ] Run the GETH

    Enter in a empty folder (really) and you can run:

    ```command
    nohup geth --goerli --cache 4096 --http --http.addr INTERNAL_IP
    --http.corsdomain *--http.vhosts* --http.api
    "admin,eth,debug,miner,net,txpool,personal,web3" --ws --ws.addr INTERNAL_IP
    --graphql --graphql.corsdomain *--graphql.vhosts* --vmdebug --metrics >
    ~/temp/logs/geth-logs.log &
    ```

- [ ] Wait the geth command finished to move to next step.

## PostgreSQL

- [ ] [Install](https://www.postgresql.org/download/) PostgreSQL
- [ ] Create a PostgreSQL User to project.

    ```command
    createuser --interactive
    ```

- [ ] Create a database

    ```command
    createdb hermez-db
    ```

- [ ] Create a password to User created at step 2

    ```command
    ALTER USER hermez WITH PASSWORD 'Your#Password';
    ```

- [ ] Test connection on localhost with all steps made until now

    ```command
    psql --host=localhost --dbname=hermez-db --username=db-username
    ```

- [ ] If you need expose the database:

    In the `/etc/postgresql/<version>/main/postgresql.conf` file, change the value of `listen_address` to your host IP

    ```command
    listen_address = "1.2.3.4" #
    ```

## Nginx

- [ ] [Install](https://www.nginx.com/resources/wiki/start/topics/tutorials/install/) Nginx

- [ ] We need create configuration file to Hermez Node with this content:

    File:

    ```command
    /etc/nginx/conf.d/hermeznode.conf
    ```

    Content:

    ```txt
    server {
        listen 80;
        listen [::]:80;

        server_name EXTERNAL_IP;

        location / {
            proxy_pass http://localhost:8086/;
        }
    }
    ```

    This nginx configuration will help us to expose the hermez node.

## Hermez-Node

- [ ] Clone Hermez Node project and do a `make` command into project folder.

    ```command
    git clone git@github.com:hermeznetwork/hermez-node.git $GOPATH/src/github.com/hermeznetwork/hermez-node
    cd hermez-node
    make
    ```

- [ ] Add `dist` path generated after `make` command inside of Hermez Node
folder.

    ```command
    export PATH=$PATH:$GOPATH/src/github.com/hermeznetwork/hermez-node/dist
    ```

- [ ] Copy the config file from `cmd/heznode/cfg.builder.toml` to
`localconfig/cfg.builder.toml` and change some values:
  - Use `0.0.0.0` instead of `localhost` to [API.Address] value

    Is so that the API can be accessed by users outside the host.

  - Add APIKey to `PriceUpdater.Fiat` section.
      Get a key [here](https://exchangeratesapi.io/)

  - Change PostgreSQL section with the values of user created before.
    - HostWrite:  Use the IP from server instead of localhost

  - Use the IP from host instead of localhost to [Web3.URL], the port is same.
    Example:

    ```text
    URL = "IP:8545"
    ```

  - In the [SmartContracts] section:
    At Rollup, change to `0xf08a226B67a8A9f99cCfCF51c50867bc18a54F53`. 
    This is the address of Smart Contract used on Goerli.
  - In the [Coordinator] section:
    - At ForgerAddress, you need add the your metamask address.
    - At ProofServerPollInterval, change to `3s`.
    - At SyncRetryInterval, change to `2s`.
    - At ForgeNoTxsDelay, change to `300s`.
  - In the [Coordinator.FeeAccount] section
    - At Address, you need add the your metamask address.
  - In the [Coordinator.EthClient] section
    - At MaxGasPrice, change to `12500000000000000000`
  - In the [Coordinator.Etherscan] section
    - Use this value to APIKey: `Insert an Etherscan key`.

- [ ] Run the hermez-node as sync mode

    ```command
    nohup heznode run --mode sync --cfg /home/youruser/go/src/github.com/hermeznetwork/hermez-node/localconfig/cfg.buidler.toml > ~/temp/logs/hermez-node.log 2>&1 &
    ```
