## Test Prover

### Server Proof API
It is necessary to have a docker with server locally.

The instructions in the following link can be followed:
https://github.com/hermeznetwork/test-info/tree/main/cli-prover

> It is necessary to consult the pre-requirements to follow the steps of the next summary

A summary of the steps to follow to run docker would be:

- Clone the repository: https://github.com/hermeznetwork/test-info
- `cd cli-prover`
- `./cli-prover.sh -s localhost -v ~/prover_data  -r 22`
- To enter docker: `docker exec -ti docker_cusnarks bash`
- Inside the docker: `cd cusnarks; make docker_all FORCE_CPU=1`
- Inside the docker: `cd config; python3 cusnarks_config.py 22 BN256`
- To exit docker: `exit`
- Now, the server API can be used. Helper can be consulted with: `./cli-prover.sh -h`
- Is necessary to initialize the server with: `./cli-prover.sh --post-start <session>`
- When `./cli-prover.sh --get-status <session>` is `ready` can be run the test.

> The session can be consulted with `tmux ls`. The session will be the number of the last session on the list.

### Test

`INTEGRATION=1 go test`