# Example 2
Example 2 is a basic api, pack setup with criteria added to the flow.

#### Perform the set-up

```
cd $GOPATH/src/github.com/HotelsDotCom/flyte/examples/example2
docker-compose up
```

### MessageSender Pack
The messagePack is a very basic pack that'll allow you to send messages (from the commandline) to the flyte api, these messages
can then be used by the flows to interact with.

#### MessageSender Pack set-up
Launch the messagepack pack by opening a new terminal and using
```
cd $GOPATH/src/github.com/HotelsDotCom/flyte/examples/messagesender
go build
export FLYTE_API="http://localhost:8080"
 ./messagesender
```

Check that the packs have deployed...

```
curl -X GET http://localhost:8080/v1/packs -H 'accept: application/x-yaml'
```

From the logs:
```
api_1     | [INFO] 2017/11/07 13:53:31 handler.go:114: Pack registered: PackId=MessageSenderPack
```

### Loading Flows
flow2 is flow1 with criteria added [Criteria](https://github.com/HotelsDotCom/flyte/blob/master/README.md#criteria), again you can view your flow in the browser [flow2](http://localhost:8080/v1/flows/flow2)

```
curl -v -X POST http://localhost:8080/v1/flows -H 'content-type: application/x-yaml' -T flow2.yaml
```

```
api_1     | [INFO] 2017/10/30 11:00:59 handler.go:94: Flow added: flowId=flow2
```

In the second flow, we've added the following criteria ```"criteria": "{{ Event.Payload|match:'^hello hello$' }}"```.To show how this functions on the first messagepack prompt type...

### Sending Messages
Go to the terminal where you launched the first MessageSender Pack, and enter the following


```
deploy app1.0
```

Output from docker-compose logs
```
api_1     | [INFO] 2017/11/29 11:57:59 handler.go:105: Received ActionEvent: Name=Output Pack={Id:Shell Name:Shell Labels:map[]}
api_1     | [INFO] 2017/11/29 11:57:59 command.go:77: Action set to SUCCESS: ActionId=5a1ea0c7713559000130bd1d CommandName=Shell PackName=Shell PackLabels=map[]
api_1     | [INFO] 2017/11/29 11:58:36 handler.go:124: Flow deleted: flowId=flow2
api_1     | [INFO] 2017/11/29 11:58:36 handler.go:94: Flow added: flowId=flow2
api_1     | [INFO] 2017/11/29 11:58:42 handler.go:75: Received Event: EventName=SendMessage Pack={Id:MessageSender Name:MessageSender Labels:map[]}
api_1     | [INFO] 2017/11/29 11:58:42 step.go:89: EventName=SendMessage matched Flow=5a1ea0f2713559000130bd1f FlowDefId=flow2 Step=output
api_1     | [INFO] 2017/11/29 11:58:42 step.go:185: ActionResponse set to NEW: ActionId=5a1ea0f2713559000130bd20 CommandName=Shell PackName=Shell PackLabels=map[] Step=output Flow=5a1ea0f2713559000130bd1f FlowDefId=flow2 input=echo -e command: 'deploy app1.0'
api_1     | [INFO] 2017/11/29 11:58:44 command.go:88: ActionResponse set to PENDING: ActionId=5a1ea0f2713559000130bd20 CommandName=Shell PackName=Shell PackLabels=map[]
shella_1  | 2017/11/29 11:58:44 script: "echo -e command: 'deploy app1.0'" -> command: deploy app1.0
api_1     | [INFO] 2017/11/29 11:58:44 handler.go:105: Received ActionEvent: Name=Output Pack={Id:Shell Name:Shell Labels:map[]}
api_1     | [INFO] 2017/11/29 11:58:44 command.go:77: Action set to SUCCESS: ActionId=5a1ea0f2713559000130bd20 CommandName=Shell PackName=Shell PackLabels=map[]
```

You can see that the event was picked up and went through three states, "NEW", "PENDING" and "SUCCESS". You can view this transaction in the browser by getting the
FlowId ```<flow-id>``` in this example it'll be different when you run it and viewing it here: http://localhost:8080/v1/flow-executions/<flow-id>

### Clean Up
kill all the sessions you have running! and then ```docker-compose rm -rf```
