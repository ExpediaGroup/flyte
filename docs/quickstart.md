# Quick start

## Writing Flows

The main interaction of users with flyte is in writing flows for flyte to execute.
 
Flows are a list of steps that define a particular use case - for example triggering the deploy of an app when a user
types the message "deploy foo-app 1.2.0" in a particular chat room.
Each step in a flow consists of an event that triggers it (e.g. an instant message being observed in a particular chat
room); criteria that must be satisfied for the step to run (e.g. the message matches a certain regex); and finally an
action that will be executed off the back of the step (e.g. triggering the deployment system to deploy the requested app).

flyte packs are self-contained apps that are responsible for executing these actions and sending events to the flyte api.
Packs are domain specific and new ones can be created as and when required. For example the bamboo pack can be used to
trigger bamboo builds and will send events to the flyte api to inform it of build successes, failures etc. Flow writers
can then look out for these events in their flow.

The happy day flow for the above deploy scenario would look to an end user something like this:

1. A user types "add label staging to build 'abc-1.2.0" in the "buildRoom"
1. The app responsible for labeling build (in this case 'Bamboo') is triggered to label the specified app
1. A instant message is sent back to the "deployRoom" chat room notifying the requester that the labeling was successful
  
3 components are in play here: 

 - the "Slack" pack that:
    1. Will send events to flyte when it observes messages being sent and
    2. Exposes a command that can be called to send a message programmatically
 - the "Bamboo" pack that will expose a command that can be called to trigger labeling of an app and will send an event
 to flyte when this is complete
 - flyte itself that will handle the interactions with the packs and execute the deploy flow that the user defines.
  
The above 'deploy' flow would be defined in flyte as follows:

## Define a flow

For more information, check [flows](flows.md) page.

## Create an item in the datastore



    curl -v -X PUT -F "description=teams.json" -F "value=@teams.json;type=application/json" http://localhost:8080/v1/datastore/teams.json

For more information, check [datastore](datastores.md) page.

## Installing a flow

Now that our flow is ready, we only need to push our flow to Flyte:

    curl -v -X POST http://localhost:8080/v1/flows -H 'content-type: application/x-yaml' -T flow.yaml

## What’s Next?

- Read [flow page](flows.md) to get more insides about flow creation.  
- Read [datastore page](datastores.md).
- Checkout [pack page](packs.md) if you want to create your own pack.  