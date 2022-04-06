[![Build Status](https://github.com/ExpediaGroup/flyte/workflows/Build/badge.svg?branch=master&event=push)](https://github.com/ExpediaGroup/flyte/actions?query=workflow:"Build"branch:"master")
[![Docker Stars](https://img.shields.io/docker/stars/expediagroup/flyte.svg)](https://hub.docker.com/r/expediagroup/flyte)
[![Docker Pulls](https://img.shields.io/docker/pulls/expediagroup/flyte.svg)](https://hub.docker.com/r/expediagroup/flyte/)

<p align="center">
  <img align="center" src="https://github.com/ExpediaGroup/flyte/blob/master/docs/images/flyte_logo_compact.png" width="300">
</p>

## Table of contents
* [Overview](#overview)
* [Getting Started](#getting-started)
* [Running](#running)
* [Running the tests](#running-the-tests)
* [More resources](#more-resources)
* [Contributing](#contributing)
* [License](#license)

## Overview

Flyte binds together the tools you use into easily defined, automated workflows. It is a request-based, decoupled automation engine which allows you to define flows to enable integrated behaviour across these disparate tools.

Flyte has chat-ops enabling integrations for Slack, as well as some other out-of-the-box [integrations](https://github.com/ExpediaGroup?utf8=%E2%9C%93&q=flyte+pack&type=&language=). These integrations, or packs can be added to and extended easily by using Flyte's RESTful API.

Some of the applications already in use include chat-ops based inventory management, host/container administration and orchestration, and deployment of applications into Kubernetes.

Automation is done using flows which essentially take form of "if this happens in system A, then do this in system B". For example you could create a flow that triggers sending an email to a team's email if their app's deployment has failed, or create a **[deployment pipeline triggered from an instant chat message](docs/images/flow.gif)**.

### How it works

The Flyte-API acts as the orchestrator and is backed by a (MongoDB) database server in which Flows, Packs references and Datastore entries are stored. Before continuing with the high level architecture diagram, lets introduce some of the key concepts of Flyte:

- **Flows** are a list of steps that define a particular use case - for example triggering the deploy of an app when a user types the message "deploy foo-app 1.2.0" in a particular chat room. Each step in a flow consists of:
    - An **Event** that triggers it (e.g. an instant message being observed in a particular chat room).
    - A **Criteria** that must be satisfied for the step to run (e.g: message matches certain regex).
    - An **Action** to be executed if the criteria matches.

    You can find more info about flows [here](docs/quickstart.md).
    
- **Packs** are self-contained apps that are responsible for executing flow actions and sending events to the flyte-api. For instance, [flyte-slack-pack](https://github.com/ExpediaGroup/flyte-slack) consumes events/messages from a slack channel but also sends messages via Slack. 
- **DataStores** are basically configuration properties shared between flows. For instance, list of environments, urls, etc.

![component diagram](docs/images/component_diag.png)

#### How are flows executed
 
1. Packs consume events from external services - for example, a Slack message from a specific channel.
1. Packs transform these external events to flyte events before pushing them to Flyte-API.
1. Flyte-API receives an event and triggers any flow which is listening to that event.
1. This flow execution will create an action and flyte-api will store it in its database.
1. Packs will poll for new actions to Flyte-API.
1. Flyte-API will assign an action to a pack.
1. Pack will execute the action and return the result as a new event, which will trigger the next step in our current flow or a different one.


![component diagram](docs/images/api_to_pack.png)  

>_Flyte-API / Flyte Pack interaction: The Flyte Pack queries Flyte-API which returns a response._ 

## Getting Started
   
Check out the [Quick Start](docs/quickstart.md) documentation to get started on building new flows or custom packs.

## Running

The are a number of ways to run flyte and its mongo db.
Note that the default mongo host and port for flyte is `localhost:27017` (this value can be changed using the
`FLYTE_MGO_HOST` env variable).

The port number can be overridden, see [Configuration](docs/configuration.md#port-configuration)

Once running, flyte will be available on http://localhost:8080 (TLS disabled), or http://localhost:8443 (TLS enabled).
  


#### From Source

Pre-req: must have [modules](https://github.com/golang/go/wiki/Modules) installed and Go 1.11 or higher.

Build & run:

```
make run
```

Command starts mongo docker container and flyte go executable in the background. Output is written to `flyte.out` file.

or manually...

```
go test ./... -tags="integration acceptance" //remove tags if only want to run unit tests
docker run -d -p 27017:27017 --name mongo mongo:latest
go build && ./flyte
```

#### Using Docker

```
make docker-run
```

or manually...


```
docker run -d -p 27017:27017 --name mongo mongo:latest
docker build -t flyte:latest .
docker run -p 8080:8080 -e FLYTE_MGO_HOST=mongo -d --name flyte --link mongo:mongo flyte:latest
 
```

## Running the tests

### Acceptance Tests

For acceptance tests:

```bash
go test ./... -tags=acceptance
```

These will start a disposable docker mongo container and flyte on randomly available TCP ports. If mongo can't be started (e.g docker is not available in the path),
tests will be skipped and won't fail the build.

Acceptance tests will not run when building flyte in docker.

The tests can be run in an IDE by running the test suite in "acceptance_test.go".

### Integration Tests

For integration tests:

```bash
go test ./... -tags=integration
```

For the mongo db integration tests, which are slower running:

```bash
go test ./... -tags=db
```

To run both:

```bash
go test ./... -tags="integration db"
```

Please note that both unit and integration tests will run using the above command/s.

### Postman

There are a number of postman files in [postman](postman) folder that can be used to test running flyte

## More resources
- [Security](docs/security/security.md)
- [Audit](docs/audit.md)
- [Configuration](docs/configuration.md)
- [Flows](docs/flows.md)
- [Packs](docs/packs.md)
- [DataStores](docs/datastores.md)

## Contributing

Please read [CONTRIBUTING.md](docs/contributing/overview.md) for details on our code of conduct, and the process for submitting pull requests to us.

## License
This project is licensed under the Apache License License - see the [LICENSE](LICENSE) file for details

