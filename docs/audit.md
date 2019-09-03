## Flow Executions Audit
Flyte provides easy way to audit flow executions. To search for flow executions use this request
```
/v1/audit/flows
```

This will return latest 50 flow executions. You can filter all flow executions by providing additional request parameters:
- flowName | flow name
- stepId | step id
- actionName | command name
- actionPackName | command pack name
- actionPackLabels | command pack labels as comma delimited string of key value pairs eg. env:staging,foo:bar
- start | start index, could be used for pagination, default is 0
- limit | number of results, default value is 50

You can also view directly individual flow execution if you know id. For example:
```
/v1/audit/flows/5ab24a266f42ed00054733d9
```

### Data TTL

By default the action collection data will expire after a year. To change this, set the following env variable:

- `FLYTE_TTL_IN_SECONDS`
