## Notes
- Access to AWS to download docker containers
  - https://www.notion.so/AWS-info-0f116600256c4370b8d9f2be47f961ba
- The very first time to run it, it takes a while since docker is downloading all proper docker images
  - **IMPORTANT NOTICE**: there is the line in the script, that deletes every docker image, container and network on the machine. 
  If you don't want it, delete or comment this line in the script - https://github.com/hermeznetwork/integration-testing/blob/main/scripts/it_local_launcher.sh#L60
- Ports that should available
  - https://github.com/hermeznetwork/integration-testing#issues
  - careful with docker and VPN, database port could also break if you have automatically running postgresDb on your machine
- See docker containers running in your machine:
  - `docker ps`
- See docker container logs in runtime:
  - `docker logs -f #docker_name`
  - Example: `docker logs -f docker_hermez-node_1`
- See logs file generated in: https://github.com/hermeznetwork/integration-testing/tree/main/test#logs-tests

# Run specific hermez-node on integration-testing

#### Requirements

- Tools: docker, docker-compose, npm v7.0+, nodejs v14.0+, aws cli 2.0+, git, jq
- Set up as env vars AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY with vars 
from paragraph *To push/pull images (hermez-docker user)*
from there https://www.notion.so/AWS-info-0f116600256c4370b8d9f2be47f961ba
- Make sure you have all requirements set up, including AWS variables

##### Run those commands to launch integration test:
```
git clone git@github.com:hermeznetwork/integration-testing.git
cd integration-testing
npm i
make build-hermezjs
npm run pretest

// change commit hermez-node in https://github.com/hermeznetwork/integration-testing/blob/main/Makefile#L8

make build-hermez-node
mkdir log
DEV_PERIOD=3 MODE=coord npx mocha ./test/**/*.mochatest.js --exit > "log/report.txt" 2>&1
```