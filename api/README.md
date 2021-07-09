# Hermez API

Easy to deploy and scale API for Hermez operators.
You will need to have [docker](https://docs.docker.com/engine/install/) and [docker-compose](https://docs.docker.com/compose/install/) installed on your machine in order to use this repo.

## Documentation

As of now the documentation is not hosted anywhere, but you can easily do it yourself by running `./run.sh doc` and then [opening the documentation in your browser](http://localhost:8001)

## Mock Up

To use a mock up of the endpoints in the API run `./run.sh doc` (UI + mock up server) or `./run.sh mock` (only mock up server). You can play with the mocked up endpoints using the [web UI](http://localhost:8001), importing `swagger.yml` into Postman or using your preferred language and using `http://localhost:4010` as base URL.

## Editor

It is recommended to edit `swagger.yml` using a dedicated editor as they provide spec validation and real time visualization. Of course you can use your prefered editor. To use the editor run `./run.sh editor` and then [opening the editor in your browser](http://localhost:8002).
**Keep in mind that you will need to manually save the file otherwise you will lose the changes** you made once you close your browser seshion or stop the server.

**Note:** Your browser may cache the swagger definition, so in order to see updated changes it may be needed to refresh the page without cache (Ctrl + Shift + R).
