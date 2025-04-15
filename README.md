# SmsService

### How it works from kafka point of view

Kafka producer
```
                                  "smstoauth" topic
	+--------------+                                      +-------------+
        | Auth Service | <- Consume   [sms code]   <- Produce | Sms Service |
	+--------------+                                      +-------------+
```

Kafka consumer
```
                                  "authtosms" topic
	+--------------+                                      +-------------+
	| Auth Service | Produce -> [phone number] Consume -> | Sms Service |
	+--------------+                                      +-------------+
```

### Running docker images

Run zookeeper docker images:
```
sudo docker compose up zookeeper -d
```

Run kafka docker images:
```
sudo docker compose up kafka-ui kafka-1 -d
```

See more details in the docker-compose file.
