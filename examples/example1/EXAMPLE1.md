# Example 1
Example 1 is a basic api, pack and flow set-up.

#### Perform the set-up

```
cd $GOPATH/src/github.com/HotelsDotCom/flyte/examples/example1
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
flow1 is a basic flow that'll look for any events from the MessageSender pack, with the event name SendMessage and will utilise the
shell pack that was deployed using docker-compose, the flow will tell the shell pack to echo out a message.

```
curl -v -X POST http://localhost:8080/v1/flows -H 'content-type: application/x-yaml' -T flow1.yaml
```

You can see from the below log that the flow has now registered, you can also view the flow [flow1](http://localhost:8080/v1/flows/flow1)
```
api_1     | [INFO] 2017/10/30 11:00:59 handler.go:94: Flow added: flowId=flow1
```


### Sending Messages
Go to the terminal where you launched the first MessageSender Pack, and enter the following

```
hello
```

Output from docker-compose logs

```
api_1     | [INFO] 2017/11/28 14:20:04 handler.go:75: Received Event: EventName=SendMessage Pack={Id:MessageSender Name:MessageSender Labels:map[]}
api_1     | [INFO] 2017/11/28 14:20:04 step.go:89: EventName=SendMessage matched Flow=5a1d70941a037200015eba11 FlowDefId=flow1 Step=output
api_1     | [INFO] 2017/11/28 14:20:04 step.go:185: ActionResponse set to NEW: ActionId=5a1d70941a037200015eba12 CommandName=Shell PackName=Shell PackLabels=map[] Step=output Flow=5a1d70941a037200015eba11 FlowDefId=flow1 input=echo -e  this is the payLoad: hello this is the packName:  MessageSender
api_1     | [INFO] 2017/11/28 14:20:06 command.go:88: ActionResponse set to PENDING: ActionId=5a1d70941a037200015eba12 CommandName=Shell PackName=Shell PackLabels=map[]
shella_1  | 2017/11/28 14:20:06 script: "echo -e  this is the payLoad: hello this is the packName:  MessageSender" -> this is the payLoad: hello this is the packName: MessageSender
api_1     | [INFO] 2017/11/28 14:20:06 handler.go:105: Received ActionEvent: Name=Output Pack={Id:Shell Name:Shell Labels:map[]}
mongo_1   | 2017-11-28T14:20:06.089+0000 I NETWORK  [thread1] connection accepted from 172.18.0.3:53396 #6 (2 connections now open)
api_1     | [INFO] 2017/11/28 14:20:06 command.go:77: Action set to SUCCESS: ActionId=5a1d70941a037200015eba12 CommandName=Shell PackName=Shell PackLabels=map[]
```

You can see that the event was picked up and went through three states, "NEW", "PENDING" and "SUCCESS". You can view this transaction in the browser by getting the
FlowId ```<flow-id>``` in this example it'll be different when you run it and viewing it here: http://localhost:8080/v1/flow-executions/<flow-id>

### Clean Up
kill all the sessions you have running! and then ```docker-compose rm -rf```
