# Flows

## What is a Flow?

Flows are a list of steps that define a particular use case - for example triggering the deploy of an app when a user
types the message "deploy foo-app 1.2.0" in a particular chat room.

Each step in a flow consists of an event that triggers it (e.g. an instant message being observed in a particular chat
room); criteria that must be satisfied for the step to run (e.g. the message matches a certain regex); and finally an
action that will be executed off the back of the step (e.g. triggering the deployment system to deploy the requested app).


## Flow Structure

Flows are defined in YAML (or JSON). Flow definition structure, as well as required and optional elements are listed below:

    name: "flow_name"                                        # required
    description: "flow description."                         # optional
    steps:                                                   # optional
      - id: "step id"                                        # optional
        criteria: "{{ Event.Payload|match:'^something' }}"   # optional
        context:                                             # optional
            key: value            
        dependsOn:                                           # optional
          - "flow_step_id"
        event:                                               # required
            packName: "pack_name"                            # required
            name: "event_name"                               # required
            packLabels:                                      # optional
                key: value
        command:                                             # required
            packName: "pack_name"                            # required
            name: "command_name"                             # required
            packLabels:                                      # optional
                key: value
            input: 'echo -e  this is the payLoad: {{  Event.Payload }} this is the packName:  {{ Event.Pack.Name }}'

The generic form of a flow is:

- The name of the flow.
- The description of the flow.
- A list of steps that define the current flow, consisting of:
    - An ID that will help to define dependencies between steps of a flow if needed.
    - The [criteria](#Criteria-Comparison) to match to trigger the step.
    - A [context](#Context) consisting of string key/value pairs that is persisted across the flow. 
    - A list of step ids that the current step [depends on](#DependsOn).
    - The command to execute when the criteria is matched, consisting of:
        - The name of the pack where the command belongs.
        - The name of the command to execute.
        - The map of labels that a pack must match to execute this command.
        - An object containing all the required input data to execute the pack command.
    - The event that will trigger this step, consisting of:
        - The name of the pack that the event came from.
        - The name of the incoming event.
        - The map of labels of the pack that the event came from

### Context

Every flow has a context which is a map of string key/value pairs that is persisted across the flow. 
Users can populate values in the context in one step and refer to them in another. The context therefore builds up
across the steps of a flow as values are added.
For example the useful parts of an event payload in one step can be saved and referred to in a later step.
The context can be thought of as session storage for the lifespan of the flow. 
 
The context is the first thing to be evaluated in a step's execution - this means that values from the context can be
used in the event packLabels matching, the criteria and command of the very same step. A useful side effect of this is that the expression to extract a value doesn't have
to be repeated multiple times in the same step - the value can be extracted, assigned to a context value, and then the
context value used from thereon out.

### Pack Labels

Pack labels can be used in 2 places in a step - in the incoming event and in the outgoing command. 
In both cases they are used as a filter, allowing a user to be as specific or as general as they want about what packs
can trigger the step and what packs can handle actions from the step.

* For a pack's event to trigger a step then the pack that sent the event must match the packName and include ALL the
packLabels defined in the step's event.
* For a pack to pick up and execute a step's command then the pack must match the packName and include ALL the
packLabels defined in the step's command.

To see what pack labels a pack is defined with, then you can look at its pack definition under `/packs/<packId>`
 
For example take a flow with the event definition below that forms part of a step: 

```
....
event:                                               
    packName: "Bamboo"
    name: "BuildSuccessful"
    packLabels:
        "env" : "staging1"
        "network" : "lab"
....
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

### DependsOn

There are broadly 2 types of steps in a flow:

1. Those that are 'entry points' to the flow that you want to be triggered by events that originated from outside the
flow. These can be thought of as trigger steps.
2. Those that are mid-flow and internal to the flow that should only be triggered by events that originated from within
the flow.

```
    .....
    steps:                                                   
      - id: "slack_start_to_bamboo"
        criteria: "{{ Event.Payload.channelId == 'XYZ12345' && Event.Payload.Message|match:'^add label$'}}",                                       // the criteria that must be met for this step to execute, i.e. is correct room and message is correct format                                        
        context:                                                    // extracts details from the event payload and adds to the context - this is availble to refer to in the same step and across steps in a flow             
            label: "{{ Event.Payload.Message|split:' '|index:3 }}"
            build: "{{ Event.Payload.Message|split:' '|index:6 }}"
            requestor: "{{ Event.Payload.User }}"
            Room: "{{ Event.Payload.channelId }}"            
        event:                                                      // if flyte receives an event of this type then it will trigger this step (assuming criteria is met)                      
            packName: "Slack"                            
            name: "MessageReceived"                               
            packLabels:                                      
                env: "staging"
        command:                                             
            packName: "Bamboo"                            
            name: "AddLabel"                             
            input: 
                build: "{{ Context.build }}"
                label: "{{ Context.label}}"

      - id: "bamboo_to_slack"
        dependsOn:                                           
            - "slack_start_to_bamboo" // indicates that step with id "slack_start_to_bamboo" must have been previously executed before this step will run
        event:                      
            packName: "Bamboo"                            
            name: "LabelAdded"                               
        command:                                             
            packName: "Slack"                            
            name: "SendMessage"
            packLabels:                                      
                env: "staging"                             
            input:
                channelId: '{{ Context.Room }}'
                message: "Hi {{ Context.requestor }} - labeled {{ Context.build}} build as {{ Context.label }}"
```

For example in the above flow the "slack_start_to_bamboo" step is the entry point into our flow - we want it be
triggered from any slack event that matches the criteria.
Conversely the step "bamboo_to_slack" is internal to the flow - we don't want it being triggered by any old
Bamboo.LabelAdded event - we only want it to be triggered if the event is off the back of the Bamboo.AddLabel command we
executed in the first step.

This is where `dependsOn` and `id` come into play. The dependsOn clause marks a step as internal and means that at least
one of the steps listed in the dependsOn must have previously been executed in the flow. 

So in the above example the 2nd step has a dependsOn clause on the first ("slack_start_to_bamboo"). This marks the
2nd step as internal and it will only be triggered when:

1. flyte receives an Bamboo.LabelAdded event, **AND**
1. the Bamboo.LabelAdded event is a response to the Bamboo command the flow previously triggered, **AND**
1. the "slack_start_to_bamboo" step has previously been triggered

Without a dependsOn clause, a step is a 'trigger step' and can be triggered by any matching event.
Ids just need to be unique within a flow. The dependsOn does not have to refer to the immediate previous step - it can be
any set of steps that is a prerequisite for the current step.

## Templating

Templates can be used at numerous points to define dynamic values in the flow definition. 
The templates use [Pongo](https://github.com/flosch/pongo2) which is a Golang implementation of Django templates.

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


There are a number of [custom functions](https://github.com/HotelsDotCom/flyte/blob/master/template/pongo.go#L41) available to use in templates:

*   `randomInt(upperBound)` - creates a random integer in the range [0, upperBound)
*   `randomAlpha(length)` - creates a random string of the specified length containing the characters \[A-Za-z]
*   `base64Encode(string)` - base 64 string encoding
*   `base64Decode(string)` - base 64 string decoding
*   `datastore(key)` - this is a function that extracts values from the flyte [datastore](datastores.md).
*   `template(template, context)` - this function resolves a pongo template (first argument) using the provided context
    (second argument). The context has to be of type `map[string]interface`


## Criteria Comparison

The criteria should adhere to pongo template language and should evaluate to true or false. 

There are a number of [custom filters](https://github.com/HotelsDotCom/flyte/blob/master/template/pongo.go#L34) available to use:

*   `kvp` - parses comma separated key=value pairs from a single string piped to the filter into a map\[string]string
*   `key` - retrieves the specified element from a piped in map
*   `index` - retrieves the specified element from a piped in slice
*   `match` - returns a boolean as to whether the piped in data meets the provided regex
*   `matchesCron` - returns a boolean if the piped in data is a time in RFC3339 format which matches the
    [cron expression](https://en.wikipedia.org/wiki/Cron#CRON_expression) argument.

Pongo2 provides a number of [inbuilt filters](https://github.com/flosch/pongo2/blob/master/filters_builtin.go#L3) in addition to these.

For examples please see [template/template_test.go](../template/template_test.go)

## Variable Interpolation

Flyte uses Jinja templating syntax to have access to data stored in the context or the event that triggered a command.

    command:
        packName: Slack
        name: SendMessage
        input:
          channelId: "{{ Context.Room }}"
          threadTimestamp: "{{ Context.Tts }}"
          message: "Consider it done!"
          
## Installing a new flow

You can easily install new flows to Flyte by using its REST API:

    curl -v -X POST http://localhost:8080/v1/flows -H 'content-type: application/x-yaml' -T flow.yaml


where `flow.yaml` is the file where your flow definition is stored.

## Examples

- Simple flow. [code](../examples/example1) 
- Simple flow with criteria. [code](../examples/example2)
- Flow with multiple steps, criteria, datastore and labels. [code](../examples/example3)