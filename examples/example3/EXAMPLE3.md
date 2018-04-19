# Example 3

#### Perform the set-up

```
cd $GOPATH/src/github.com/HotelsDotCom/flyte/examples/example3
docker-compose up
```

### MessageSender Pack
The messagePack is a very basic pack that'll allow you to send messages (from the commandline) to the flyte api, these messages
can then be used by the flows to interact with.

#### MessageSender Pack set-up
Launch the messagepack pack by opening a new terminal and using
```
cd $GOPATH/src/github.com/HotelsDotCom/flyte/examples/messagesender
dep ensure
go build -tags=examples
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

flow3 is flow1 and flow2 with crieria and labels [Pack Labels](https://github.com/HotelsDotCom/flyte/blob/master/README.md#pack-labels), again you can view your flow in the browser [flow3](http://localhost:8080/v1/flows/flow3)

```
curl -v -X POST http://localhost:8080/v1/flows -H 'content-type: application/x-yaml' -T flow3.yaml
```

```
api_1     | [INFO] 2017/10/30 11:00:59 handler.go:94: Flow added: flowId=flow3
```

## Datastore
flyte provides a datastore that allows reference data to be persisted and made available for use in flow definitions. The datastore data is global and items are added by POSTing a multipart request to its resource
### Load Datastore
```
curl -v -F "key=teams" -F "description=hipchat teams" -F "value=@datastore/teams.json;type=application/json" http://localhost:8080/v1/datastore
```

from the logs:

```
api_1     | [INFO] 2017/10/30 11:00:59 handler.go:109: Data item stored: key=teams contentType=application/json
```

Check that the datastore has loaded correctly..

```
curl -X GET http://localhost:8080/v1/datastore/teams
```
#### Output
```javascript
{
    "devinf": {
        "email": "devinf@example.com",
        "hipchat_room": "abcdef"
    },
    "devs": {
        "email": "devs@example.com",
        "hipchat_room": "10001"
    }
}
```



In the final flow, we expand on both flow1 and 2 by adding datastore, and a extra step. Go to the terminal with the 2nd pack running.
The one with the ABC=123 label we applied.

Then enter...

```
email 10000
```
You can see from the output below, that this entry will be picked up by the 2nd pack only and executes two steps.

```
api_1     | [INFO] 2017/11/29 11:33:54 step.go:89: EventName=SendMessage matched Flow=5a1e9b22c79af80001a5afa6 FlowDefId=flow3 Step=flow3_stepone
api_1     | [INFO] 2017/11/29 11:33:54 step.go:185: ActionResponse set to NEW: ActionId=5a1e9b22c79af80001a5afa7 CommandName=Shell PackName=Shell PackLabels=map[] Step=flow3_stepone Flow=5a1e9b22c79af80001a5afa6 FlowDefId=flow3 input=echo -e 'teama@example.com'
api_1     | [INFO] 2017/11/29 11:33:54 command.go:88: ActionResponse set to PENDING: ActionId=5a1e9b22c79af80001a5afa7 CommandName=Shell PackName=Shell PackLabels=map[]
shellb_1  | 2017/11/29 11:33:54 script: "echo -e 'teama@example.com'" -> teama@example.com
api_1     | [INFO] 2017/11/29 11:33:54 handler.go:105: Received ActionEvent: Name=Output Pack={Id:Shell.ABC.123 Name:Shell Labels:map[ABC:123]}
mongo_1   | 2017-11-29T11:33:54.554+0000 I NETWORK  [thread1] connection accepted from 172.20.0.3:39432 #2 (2 connections now open)
api_1     | [INFO] 2017/11/29 11:33:54 command.go:77: Action set to SUCCESS: ActionId=5a1e9b22c79af80001a5afa7 CommandName=Shell PackName=Shell PackLabels=map[]
api_1     | [INFO] 2017/11/29 11:33:54 step.go:89: EventName=Output matched Flow=5a1e9b22c79af80001a5afa6 FlowDefId=flow3 Step=flow3_steptwo
api_1     | [INFO] 2017/11/29 11:33:54 step.go:185: ActionResponse set to NEW: ActionId=5a1e9b22c79af80001a5afa8 CommandName=Shell PackName=Shell PackLabels=map[] Step=flow3_steptwo Flow=5a1e9b22c79af80001a5afa6 FlowDefId=flow3 input=echo -e 'i am flow3 step 2'
api_1     | [INFO] 2017/11/29 11:33:55 command.go:88: ActionResponse set to PENDING: ActionId=5a1e9b22c79af80001a5afa8 CommandName=Shell PackName=Shell PackLabels=map[]
shella_1  | 2017/11/29 11:33:55 script: "echo -e 'i am flow3 step 2'" -> i am flow3 step 2
api_1     | [INFO] 2017/11/29 11:33:55 handler.go:105: Received ActionEvent: Name=Output Pack={Id:Shell Name:Shell Labels:map[]}
api_1     | [INFO] 2017/11/29 11:33:55 command.go:77: Action set to SUCCESS: ActionId=5a1e9b22c79af80001a5afa8 CommandName=Shell PackName=Shell PackLabels=map[]

```


### Clean Up
kill all the sessions you have running! and then ```docker-compose rm -rf```