Test for GiG, I did my best to complete it while following best practices

## Running

`docker compose up`

spawns all microservices occupying ports 8080-8082.
It performs unit tests and e2e tests, they exit with 0

### Core decisions

For queue communication between publish and subscribe microservices I've used rabbitmq.
I've chosen it because it is popularity and general availability.
Both microservices are written in golang as that is the language I have been doing most of the coding lately.
As an added extra I've added e2e test and simple frontend client.

### Code

#### Microservices

[pub](./pub) - publish microservice running on port 8080 (`ws://localhost:8080/ws`)

[sub](./sub) - subscribe microservice running on port 8081 (`ws://localhost:8081/ws`)

[e2e](./e2e) - end to end test which tests challenge specified funcionality of pub and sub

unit tests - are simple unit tests for written code

[client](./client) - a simple frontend client with connect/disconnect/send/recieve functionalities running on port 8082

- to connect as publisher visit [link](http://localhost:8082/static/?endpoint=ws://localhost:8080/ws)

- to connect as subscriber visit [link](http://localhost:8082/static/?endpoint=ws://localhost:8081/ws)

#### Other

both pub and sub microservices support websocket `/ws`, `/healthcheck` and `/shutdown` route.

[wait-for-it.sh](./wait-for-it.sh) acts as a blocker until given endpoints are available

[unit-test.sh](./uni-test.sh) executes unit tests in all folders except the ones excluded by grep.

#### Improvements

CI/CD could be built with steps:

    1. run unit tests
    2. run e2e tests
    3. deploy sub and pub to given target machines
        1. build binaries
        2. build docker images and tag them (docker acting as an artefact)
        3. deploy to target machines
