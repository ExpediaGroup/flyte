# Packs

## What is a Pack?

A **Pack** is a self-contained app that is responsible for executing flow actions and sending events to the flyte-api. 
For instance, [flyte-slack](https://github.com/ExpediaGroup/flyte-slack) pack consumes events/messages from a slack channel but also sends messages via Slack. 

## Discovering Packs

There are already some [packs available](https://github.com/ExpediaGroup?utf8=%E2%9C%93&q=flyte+pack&type=&language=) in Github. Check them out before creating your own one.

As a developer you can check what packs are registered in your Flyte API instance by querying your packs endpoint:

    curl -X GET http://localhost:8080/v1/packs    

## Installing a Pack

Installing a pack is simple, you only need to provide a valid Flyte API endpoint to your pack (and a JWT if [security](security/security.md) is enabled), and [flyte-client](https://github.com/ExpediaGroup/flyte-client) will automatically register your pack definition to Flyte API.

## Developing a Pack

So you need to create a new pack? Great! Here's how to do it.

Packs are self-contained apps and language agnostic, so you can write your packs using a programming language of your choice.
However, if you planning to use golang, Flyte provides a [flyte-client](https://github.com/ExpediaGroup/flyte-client) to make the writing of flyte packs simpler.

## Automatic removal of unused or 'dead' packs

If you'd like to clean up your mongo db of unused or 'dead' packs, flyte can schedule a process to do this daily.

Set the following env variables before you start flyte:

  - FLYTE_SHOULD_DELETE_DEAD_PACKS - default is false, set to true to turn this feature on.
  - FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM - specify a time of day to do your deletion in the format 'HH:MM'. If not set the default is '23:00'.
  - FLYTE_PACK_GRACE_PERIOD_UNTIL_MARKED_DEAD_IN_SECONDS - This tells the scheduler to remove packs with a 'LastSeen' date older than the value passed. The default is 604800 (one week).
  
The scheduler will start it's cleanup for the first time after midnight on the day flyte is started.


### Using flyte-client

[Flyte-client](https://github.com/ExpediaGroup/flyte-client) is a Go library designed to make the writing of flyte packs simple. 

The client handles: 

- the registration of a pack with the flyte server
- consuming and handling command actions
- send pack events to the flyte server. 

This allows the pack writer to concentrate solely on the functionality of their pack.

You can find more information and code examples [here](https://github.com/ExpediaGroup/flyte-client).


### Using a programming language of your choice

You can mimic the flyte-client library using the language of your choice by using Flyte REST API.

**IMPORTANT:** Please use flyte [HATEOAS](https://en.wikipedia.org/wiki/HATEOAS) component to interact and navigate through our API, avoid using hardcoded endpoints.

1. **Register your pack**: your code will be responsible for posting your pack definition to the flyte server. 

    Review [swagger documentation](http://localhost:8080/swagger#!/pack/registerPack) to check the contract of this endpoint.
    
1. **Consume actions**: after registering your pack, you will now start consuming actions from Flyte. These actions contain information about the command to execute in our pack. 
Packs should use a non-blocking polling mechanism to consume actions from Flyte. Flyte will NOT push any action to your pack. You will basically need to create a polling loop that will request for new actions to Flyte API every X seconds (flyte-client has a 5s polling frequency by default). 
 
    Review [swagger documentation](http://localhost:8080/swagger#!/action/takeAction) to check the contract of this endpoint.

1. **Execute action**: once a new action is fetched from Flyte, the next step will be invoke the relevant commandHandler/code associated to that consumed action. This handler will return an event that the client will then send to the flyte api by posting the result to the `action-result` endpoint. 

    Review [swagger documentation](http://localhost:8080/swagger#!/action/actionResult) to check the contract of this endpoint.

1. **Post events**: Packs can send events to flyte in 3 ways:
                    
    - The pack can observe something happening and spontaneously send an event to the flyte server. For example a chat-ops pack, may observe an instant message being sent and raise a "MessageSent" event to the flyte server.
    - A flow on the flyte server creates an action for the pack to execute. The client will poll for this action and invoke the relevant CommandHandler that the pack dev has defined. This handler will return an event that the client will then send to the flyte server. For example the same IM pack as above may have a 'sendMessage' command that would return either a 'MessageSent' or 'MessageSendFailure' event.
    - The client will produce FATAL events as a result of panic happening while handling a Command. This will be intercepted by the client and it will recover.
                    
    **IMPORTANT**: The payload of an Event object can be any valid JSON object. Flyte will not force any special contract as every pack will return its own output data.  
    
    Review [swagger documentation](http://localhost:8080/swagger#!/event/event) to check the contract of this endpoint.
