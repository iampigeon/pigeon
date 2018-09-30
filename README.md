# Pigeon

Central notification service FTW.

# HTTP

## Get Subjects
```
  Get /subjects
```

### Headers
| Name | Type | Description |
|-|-|-|-|
|X-Api-Key|string|user api key|

### Response
```Json
{
  "data": {
    "subjects": [{
      "name": "max-air-temperature",
      "channels": ["mqtt", "sms"]
    }, {
      "name": "weekly-report",
      "channels": ["sms", "push"]
    }]
  },
  "meta": {}
}
```

# TODO

- [ ] todo 1

