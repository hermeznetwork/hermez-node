# Run specifc hermez-node on integration-testing
```
Requirements: docker, docker-compose, npm v7.0+, nodejs v14.0+, aws cli 2.0+, git, jq
Set up as env vars AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY with vars 
from paragraph *To push/pull images (hermez-docker user)*
from there https://www.notion.so/AWS-info-0f116600256c4370b8d9f2be47f961ba

Run those commands to run test:
git clone git@github.com:hermeznetwork/integration-testing.git
cd integration-testing
npm i
make build-hermezjs
npm run pretest

// change commit hermez-node in https://github.com/hermeznetwork/integration-testing/blob/main/Makefile#L8

make build-hermez-node
DEV_PERIOD=3 MODE=coord npx mocha ./test/**/*.mochatest.js --exit > "log/report.txt" 2>&1
```

## Notes
- Access to aws to download dockers
    - https://www.notion.so/AWS-info-0f116600256c4370b8d9f2be47f961ba
- The very first time to run it, it take a while since docker is downloading all proper docker images
    - next times, you can skip the command `docker system prune -a -f` and it will be faster
- Ports that should available
    - https://github.com/hermeznetwork/integration-testing#issues
    - careful with docker and VPN, database port could also break if you have automatically running postgresDb on your machine
- See dockers running in your machine:
    - `docker ps`
- See docker logs in runtime:
    - `docker logs -f #docker_name`
    - Example: `docker logs -f docker_hermez-node_1`
- See logs file generated in: https://github.com/hermeznetwork/integration-testing/tree/main/test#logs-tests 