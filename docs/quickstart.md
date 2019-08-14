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

 - the "HipChat" pack that:
    1. Will send events to flyte when it observes messages being sent and
    2. Exposes a command that can be called to send a message programmatically
 - the "Bamboo" pack that will expose a command that can be called to trigger labeling of an app and will send an event
 to flyte when this is complete
 - flyte itself that will handle the interactions with the packs and execute the deploy flow that the user defines.
  
The above 'deploy' flow would be defined in flyte as follows:

```
{
    "name": "LabelBuildFlow",                               // the name of the flow
    "description": "labels an app based on hipchat msg.",   // optional description
    "labels": {                                             // optional labels to distinguish flows with the same name
            "env": "staging"
    },
    "steps": [
        {                                                   // a flow step
            "id": "hipchat_start_to_bamboo",                // id of the step
            "event": {                                      // if flyte receives an event of this type then it will trigger this step (assuming criteria is met)
                "packName": "HipChat",
                "packLabels": {
                    "env" : "staging"
                },
                "name": "MessageReceived"
            },
            "context": {                                    // extracts details from the event payload and adds to the context - this is availble to refer to in the same step and across steps in a flow
                "label": "{{ Event.Payload.Message|split:' '|index:3 }}",
                "build": "{{ Event.Payload.Message|split:' '|index:6 }}",
                "requestor": "{{ Event.Payload.User }}"
            },
            "criteria": "{{ Event.Payload.RoomId == 'deployRoom' && Event.Payload.Message|match:'^add label$'}}",                                       // the criteria that must be met for this step to execute, i.e. is correct room and message is correct format
            "command": {                                    // calls a command on the 'bamboo' pack to trigger the deploy
                "packName": "Bamboo",
                "name": "AddLabel",
                "input": {
                    "build": "{{ Context.build }}",
                    "label": "{{ Context.label}}"
                }
            }
        },  
        {                                                   // the 2nd of the 2 steps in the flow - this sends a response back to the user that initiated the labeling request
            "id": "bamboo_to_hipchat",
            "dependsOn": [                                
                "hipchat_start_to_bamboo"                   // indicates that step with id "hipchat_start_to_bamboo" must have been previously executed before this step will run
            ],
            "event": {
                "packName": "Bamboo",
                "name": "LabelAdded"
            },
            "command": {
                "packName": "HipChat",
                "packLabels": {
                    "env" : "staging"
                },
                "name": "SendMessage",
                "input": {
                    "roomId": "deployRoom",
                    "message": "Hi {{ Context.requestor }} - labeled {{ Context.build}} build as {{ Context.label }}"
                }
            }
        }
    ]
}
```

This flow consists of 2 steps - one to look out for users typing "add label" messages & to kick off labeling build,
and the 2nd to look out for the successful labeling and to send a success message back to the user.

Steps in the flow have a number of components/features. These are detailed below.

#### Templating

Templates can be used at numerous points to define dynamic values in the flow definition. 
The templates use 'Pongo' which is a Golang implementation of Django templates - for more details about pongo see:
https://github.com/flosch/pongo2.

Note that templates are case sensitive!

Templates can be used in the following places in a flow definition:

1. As values in the 'context' map
1. As values in the event's 'packLabels' map
1. As the 'criteria' value
1. As values in the command's 'packLabels' map
1. As part of the command's 'input'

The template has a few context objects provided to it that you can make use of:

* `Event` - this is the event that triggered the current step. It has the following fields
    * Event.Name - the name of the incoming event e.g. 'BuildSuccess'
    * Event.Payload - the json payload sent by the pack e.g. 'Event.Payload.foo' would return the foo element of the
    event payload
    * Event.Pack.Name - the name of the pack that the event came from e.g. 'Bamboo'
    * Event.Pack.Labels - the map of labels of the pack that the event came from e.g. 'Event.Pack.Labels.env' might
    return for example 'staging'.
    
* `Context` - this is the context that can be used to persist data between flow steps (see below). e.g. 'Context.bar'
would return the 'bar' element stored in the context.

There are a number of custom functions available to use in templates:

*   `randomInt(upperBound)` - creates a random integer in the range [0, upperBound)
*   `randomAlpha(length)` - creates a random string of the specified length containing the characters \[A-Za-z]
*   `base64Encode(string)` - base 64 string encoding
*   `base64Decode(string)` - base 64 string decoding
*   `datastore(key)` - this is a function that extracts values from the flyte datastore - see the section below.
*   `template(template, context)` - this function resolves a pongo template (first argument) using the provided context
    (second argument). The context has to be of type `map[string]interface`

There are a number of custom filters available to use:

*   `kvp` - parses comma separated key=value pairs from a single string piped to the filter into a map\[string]string
*   `key` - retrieves the specified element from a piped in map
*   `index` - retrieves the specified element from a piped in slice
*   `match` - returns a boolean as to whether the piped in data meets the provided regex
*   `matchesCron` - returns a boolean if the piped in data is a time in RFC3339 format which matches the
    [cron expression](https://en.wikipedia.org/wiki/Cron#CRON_expression) argument.

Pongo provides a number of inbuilt filters in addition to these:
https://github.com/flosch/pongo2/blob/master/filters_builtin.go#L3.

For examples please see [template/template_test.go](template/template_test.go)

#### Context

Every flow has a context which is a map of string key/value pairs that is persisted across the flow. 
Users can populate values in the context in one step and refer to them in another. The context therefore builds up
across the steps of a flow as values are added.
For example the useful parts of an event payload in one step can be saved and referred to in a later step.
The context can be thought of as session storage for the lifespan of the flow. 
In the above example the variables 'appName', 'appVersion' and 'requestor' are added to the context in the first step
(when we receive the event from hipchat), which allows us to use them in the 2nd step and send a relevant message back
to the user.
 
The context is the first thing to be evaluated in a step's execution - this means that values from the context can be
used in the event packLabels matching, the criteria 
and command of the very same step. A useful side effect of this is that the expression to extract a value doesn't have
to be repeated multiple times in the same step - the value can be extracted, assigned to a context value, and then the
context value used from thereon out.

#### Criteria

The trigger event payload can be to be interrogated to see if matches a certain criteria. Only if it does will the step
be executed. The criteria should adhere to pongo template language and should evaluate to true or false. 
For more details about pongo see https://github.com/flosch/pongo2. In our example above we check that the hipchat event
originated from the 'deployRoom' and the message body meets a certain regex. 

#### Pack Labels

Pack labels can be used in 2 places in a step - in the incoming event and in the outgoing command. 
In both cases they are used as a filter, allowing a user to be as specific or as general as they want about what packs
can trigger the step and what packs can handle actions from the step.

* For a pack's event to trigger a step then the pack that sent the event must match the packName and include ALL the
packLabels defined in the step's event
* For a pack to pick up and execute a step's command then the pack must match the packName and include ALL the
packLabels defined in the step's command.

To see what pack labels a pack is defined with, then you can look at its pack definition under `/packs/<packId>`
 
For example take a flow with the event definition below that forms part of a step: 

```
"event": {      
             "packName": "Bamboo",
             "packLabels": {
                 "env" : "staging1",
                 "network" : "lab"
             },
             "name": "BuildSuccessful"
},
```

A pack defined with `packName='Bamboo', packLabels={'env' : 'staging1', 'foo' : 'bar', 'network' : 'lab'}` sending
event `BuildSuccessful` would be eligible to trigger the step as it is defined with the same pack name and it includes
all the pack labels in its definition (notice that the pack is defined with more labels - the labels defined in the
flow must just be a subset of the pack's actual labels).

The following packs sending the same event would NOT be able to trigger the step:

`packName='Bamboo', packLabels={'env' : 'staging2', 'network' : 'lab'}` ('env' label's value doesn't match)

`packName='Bamboo', packLabels={'env' : 'staging1', 'x' : 'y'}` (doesn't have 'network' label)

`packName='NotBamboo', packLabels={'env' : 'staging1', 'foo' : 'bar', 'network' : 'lab'}` (packName is wrong)


Labels allow you to be as specific or as general as you want about what instances of a pack handle parts of your
flow - you can apply as many labels as required to target a specific pack instance or set of instances, or leave the
labels off altogether to allow any instance of the pack to handle the work.

The values of pack labels can also be parameterised - this gives a number of benefits, one of which is flow reuse. For
example you might want the same flow in dev as in production but don't want to define the flow multiple times
(with the only difference being the labels to identify the prod or test versions of the required packs).
By parameterising the labels, on your first step you can identify what env the incoming event is from, store this in
the context and then use this value in pack labels (see the postman files for an example of this)

#### dependsOn & Id

There are broadly 2 types of steps in a flow:

1. Those that are 'entry points' to the flow that you want to be triggered by events that originated from outside the
flow. These can be thought of as trigger steps.
2. Those that are mid-flow and internal to the flow that should only be triggered by events that originated from within
the flow.

For example in the above flow the "hipchat_start_to_bamboo" step is the entry point into our flow - we want it be
triggered from any hipchat event that matches the criteria.
Conversely the step "bamboo_to_hipchat" is internal to the flow - we don't want it being triggered by any old
Bamboo.LabelAdded event - we only want it to be triggered if the event is off the back of the Bamboo.AddLabel command we
executed in the first step.

This is where `dependsOn` and `id` come into play. The dependsOn clause marks a step as internal and means that at least
one of the steps listed in the dependsOn must have previously been executed in the flow. 

So in the above example the 2nd step has a dependsOn clause on the first ("hipchat_start_to_bamboo"). This marks the
2nd step as internal and it will only be triggered when:

1. flyte receives an Bamboo.LabelAdded event, **AND**
1. the Bamboo.LabelAdded event is a response to the Bamboo command the flow previously triggered, **AND**
1. the "hipchat_start_to_bamboo" step has previously been triggered

Without a dependsOn clause, a step is a 'trigger step' and can be triggered by any matching event.
Ids just need to be unique within a flow. The dependsOn doesn't have to refer to the immediate previous step - it can be
any set of steps that is a prerequisite for the current step.

#### Command

The command section of a step details what you want to execute if the step is executed. 
The command must be an available command on a pack registered with flyte.

## Datastore

flyte provides a datastore that allows reference data to be persisted and made available for use in flow definitions.
The datastore data is global and items are added by PUTting a multipart request to its resource. The value may be in any
format. You can then select and use datastore data in your flows using the `datastore` function as described previously.

#### Datastore Example

curl: `curl -v -X PUT -F "description=hipchat teams.json" -F "value=@teams.json;type=application/json" http://localhost:8080/v1/datastore/teams.json`

File content type is optional and defaults to 'text/plain; charset=us-ascii'. File key has to be `value`.

`teams` file exmaple:
```
{
    "devinf": {
        "email": "devinf@example.com",
        "hipchat_room": "10000",
        ...
    },
    "devs": {
        "email": "devs@example.com",
        "hipchat_room": "10001",
        ...
    },
    ...
}
```

We can then use this `teams` datastore item in a flow step to lookup the email address for a given team name e.g.

```
   ...
   "command" : {
      "pack-id" : "Hipchat",
      "command-name" : "SendMessage",
      "arguments" : {
         "Message" : "Thanks for contacting us! if you have any further inquires please contact {{ datastore('teams').devinf.email` }}",
         ...
      }
   },
   ...
```

This will work only if the correct content type (`application/json`) is set, otherwise the item's value will be
resolved as a string.

