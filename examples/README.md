# Flyte Example
This example is in place to help you navigate yourself around flyte, get yourself use to the api, packs, flows.

The example will show you how to interact with the api, packs and flows. And how to manipulate the flows.

## Objectives
- Start up Flyte using docker-compose
- Deploy a basic pack
- Use flows
- Use the flyte datastore

## Requirements
- Docker. [docker installation](https://docs.docker.com/engine/installation/)
- Go version 1.11+ [go installation](https://golang.org/)

## Set-up
We utilise docker-compose,see [docker-compose](https://docs.docker.com/compose/)  to run this example. This will start the api and two versions
of the flyte-shell pack, see [flyte-shell](https://github.com/ExpediaGroup/flyte-shell).

## Packs
flyte packs are self-contained apps that are responsible for executing these actions and sending events to the flyte api. Packs are domain specific and new ones can be created as and when required. For example the bamboo pack can be used to trigger bamboo builds and will send events to the flyte api to inform it of build successes, failures etc.
Flow writers can then look out for these events in their flow.

## Flows
Flows are a list of steps that define a particular use case - for example triggering the deploy of an app when a user types the message "deploy foo-app 1.2.0" in a particular chat room. Each step in a flow consists of an event that triggers it (e.g. an instant message being observed in a particular chat room); criteria that must be satisfied
for the step to run (e.g. the message matches a certain regex); and finally an action that will be executed off the back of the step (e.g. triggering the deployment system to deploy the requested app).

You can read more on writing flows here [Writing Flows](https://github.com/ExpediaGroup/flyte/blob/master/README.md#writing-flows)

## Examples

[Example1](example1/EXAMPLE1.md)

[Example2](example2/EXAMPLE2.md)

[Example3](example3/EXAMPLE3.md)

