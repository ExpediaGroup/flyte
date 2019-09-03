# Datastores

A datastore allows reference data to be persisted and made available for use in flow definitions.
The datastore data is global and items are added by PUTting a multipart request to its resource. The value may be in any
format. You can then select and use datastore data in your flows using the `datastore` function.

## Storing values

You can add new items to the datastore by PUTting a multipart request to Flyte API datastore endpoint:

    curl -v -X PUT -F "description=teams.json" -F "value=@teams.json;type=application/json" http://localhost:8080/v1/datastore/teams.json

File content type is optional and defaults to 'text/plain; charset=us-ascii'. File key has to be `value`.

`teams` file example:
```
{
    "devinf": {
        "email": "devinf@example.com",
        ...
    },
    "devs": {
        "email": "devs@example.com",
        ...
    },
    ...
}
```

## Retrieving values

We can then use this `teams` datastore item in a flow step to lookup the email address for a given team name e.g.

```
   ...
     command:
       packName: Slack
       name: SendMessage
       input:
         message: "Thanks for contacting us! if you have any further inquires please contact {{ datastore('teams').devinf.email` }}"
   ...
```

This will work only if the correct content type (`application/json`) is set, otherwise the item's value will be
resolved as a string.
