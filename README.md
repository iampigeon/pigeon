# Pigeon

Central notification service FTW.

## HTTP

## GET Subjects
```
  GET /api/v1/subjects
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
      "channels": ["mqtt", "sms", "mandrill"]
    }, {
      "name": "weekly-report",
      "channels": ["sms", "push"]
    }]
  }
}
```

## Create message
```
  POST /api/v1/messages
```

### Headers
| Name | Type | Description |
|-|-|-|-|
|X-Api-Key|string|user api key|

### Request
```Json
{
  "message": {
    "subject_name": "max-air-temperature",
    "channels": {
      "mqtt": {
        "mqtt_payload": {
          "foo": "bar",
          "skt": "uno"
        },
        "text": "some-sms-text"
      },
      "sms": {
        "phone": "+56912345678",
        "text": "some-sms-text"
      },
      "mandrill": {
        "text": "some-sms-text",
        "token": "secure-api",
        "email": "cristobal@iampigeon.com",
        "type": "???",
        "from_email": "support@iampigeon.com",
        "from_name": "support",
        "subject": "some-text",
        "html": "some-html",
        "text": "some-text",
      },
    }
  }
}
```

### Response
```Json
{
  "data": {
    "messages": [{
      "id": "id:01CRVFYX6TG8E7ETKRQ4H21S8R",
      "channel": "mqtt"
    }, {
      "id": "id:0YGFDLJUCFG8E7ETHJDHYTOSKFJ",
      "channel": "sms",
    }, {
      "error": "internal server error",
      "channel": "mandrill"
    }]
  }
}
```

# Relations

```bash

                                                     +------------+
                                                     |            |
 +---------+                  +--------------------->+ Criterias  |
 |         |                  |         1 : 1        |            |
 |  Users  |                  |                      +------------+
 |         |                  |
 +----+----+                  |
      ^             +---------+--------+             +------------+
      |             |                  |    1 : 1    |            |
      |             | SubjectsChannels +<----------->+  Channels  |
      |             |                  |             |            |
      |             +---------+--------+             +------------+
      |                       ^
      |                       | 1 : N
      |                       v
      |                +------+-----+                +------------+
      |    1 : N       |            |     1 : N      |            |
      +--------------->+  Subjects  +<-------------->+  Messages  |
                       |            |                |            |
                       +------------+                +------------+

```

# Models

## Users

|name|type|
|-|-|-|
|id|string|
|email|string|
|password|string|

### Example

```json
{
  "users": [{
    "id": "1",
    "email": "camilo@iampigeon.com",
    "password": "password",
    "api_key": "12345"
  }, {
    "id": "2",
    "email": "chelo@iampigeon.com",
    "password": "password",
    "api_key": "abcde"
  }]
}
```

## Channels

Define all pigeons services backends available to clients.

|name|type|
|-|-|-|
|id|string|
|name|string|

### Example

```json
{
  "channels": [{
    "id": "c1",
    "name": "mqtt",
    "host": "localhost:9010"
  }, {
    "id": "c2",
    "name": "sms",
    "host": "localhost:9020"
  }, {
    "id": "c3",
    "name": "mandrill",
    "host": "localhost:9030"
  }, {
    "id": "c4",
    "name": "webhook",
    "host": "localhost:9040"
  }, {
    "id": "c5",
    "name": "telegram",
    "host": "localhost:9050"
  }]
}
```

## Subjects

Define user subject

|name|type|
|-|-|-|
|id|string|
|user_id|string|
|name|string|
|webhook|string|
|channels|[]SubjectsChannels|

## Subjects Channels

Define relations of subjects and channels

|name|type|
|-|-|-|
|id|string|
|channel_id|string|
|criteria_id|string|
|criteria_custom|int64|
|callback_post_url|string|
|options|map[string]string|

### Example

```json
{
  "subjects": [{
    "id": "s1",
    "user_id": "u1",
    "name": "max-air-temperature",
    "channels": [{
      "id": "uc1",
      "channel_id": "c1",
      "criteria_id": "t5",
      "criteria_custom": 12,
      "callback_post_url": "localhost:9999",
      "options": {
        "mqtt_topic": "/some-mqtt-topic"
      }
    }, {
      "id": "uc2",
      "channel_id": "c2",
      "criteria_id": "t5",
      "criteria_custom": 60,
      "callback_post_url": "localhost:9999"
    }]
  }, {
    "id": "s2",
    "user_id": "u2",
    "name": "weekly-report",
    "channels": [{
      "id": "uc3",
      "channel_id": "c4",
      "criteria_id": "t2",
      "criteria_custom": null,
      "callback_post_url": "localhost:9999",
      "options": {
        "headers": {
          "content-type": "application/json",
          "x-api-key": "lala123"
        }
      }
    }]
  }, {
    "id": "s3",
    "user_id": "u1",
    "name": "wisebot-service-update",
    "channels": [{
      "id": "uc4",
      "channel_id": "c6",
      "criteria_id": "t3",
      "callback_post_url": null,
      "options": {
        "wg_url": "https://wg-api-production.wisegrowth.app",
        "wg_api_key": "secure_api_key"
      }
    }, {
      "id": "uc5",
      "channel_id": "c1",
      "criteria_id": "t4",
      "callback_post_url": null,
      "options": {
        "mqtt_topic": "/wisebots/:id/service-update",
        "aws_key": "some-key",
        "aws_secret_key": "some-secret-key"
      }
    }]
  }]
}
```


## Messages

|name|type|
|-|-|-|
|id|string|
|subject_id|string|
|status|string|

### Example

```json
{
  "messages": [{
    "id": "m1",
    "subject_id": "s1",
    "status": "pending"
  }, {
    "id": "m2",
    "subject_id": "s2",
    "status": "error"
  }]
}
```

## Criteria

|name|type|
|-|-|-|
|id|string|
|name|string|
|value|int64|

### Example

```json
{
  "messages": [{
    "id": "t1",
    "name": "low",
    "value": 86400
  }, {
    "id": "t2",
    "name": "mid",
    "value": 21600
  }, {
    "id": "t3",
    "name": "high",
    "value": 3600
  }, {
    "id": "t4",
    "name": "now",
    "value": 0
  }, {
    "id": "t5",
    "name": "custom",
    "value": -1
  }]
}
```

# TODO

- [X] Define statuses (ca)
- [X] Implement status behavior on message lifecycle (ca)
- [X] Define and implements mocks struct models (ca)
- [X] Implement mock to `GET /api/v1/subject` endpoint (ca)
- [X] Implement `GET /api/v1/messages/:id/status` endpoint (ca)
- [ ] Implement status method on pigeon-go client (mt) 
- [ ] Implement `POST /api/v1/messages/:id/cancel` endpoint (ca)
- [ ] Handle cancellation in scheduler.go file (ja)
- [ ] Implement rpc route for cancel message by id (ca)
- [ ] Implement cancel method on `pigeon-go` client (mt)
- [X] Define and implement `pigeon-http` channel (ca)
- [X] Add `pigeon-http` value to subject model {options: {headers}} (ca)
- [ ] Implement 'callback_post_url' using pigeon-http service inside scheduler.go file (ca)
- [X] Append user-channels `options` to content message by channel (ca)
- [ ] Implement `Subjects` table on `BoltDB` (ca)
- [ ] Implement `Subject Channels` table on `BoltDB` (ca)
- [ ] Implement `Users` table on `BoltDB` (ca)
- [X] Add 'data.json' on docker (ca)
- [X] Implement `Messages` table on `BoltDB` (ca)
- [X] Add Status on `Messages` protobuf (ca)
- [X] Implement logic to validate and get user by x-api-key header value (ca)
- [X] Add user_id to message protobuf and implement this in put scheduler method (ca)
- [ ] Define error codes in backend.go file (ja)
- [ ] Use secure connections in all grpc connections (ja)
- [X] Add criteria model (ca)
- [X] Add criteria examples to mock (ca)
- [X] Implement criteria when create new message (ca)
- [ ] Should use cron tab in criteria_value (ca)
- [ ] Implement interface for any model
- [ ] Return Subject when has create
- [ ] Implement JWT in any request

```json
{
  "message": {
    "subject_name": "max-air-temperature",
    "channels": [{
      "channel_id": "mqtt",
      "mqtt_payload": {
        "foo": "bar",
        "baz": "zar"
      }
    }, {
      "channel_id": "sms",
      "phone": "2423545353"
    }]
  }
}
```