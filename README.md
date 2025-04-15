# SmsService

### Run the application

Run zookeeper docker image:
```
sudo docker compose up zookeeper -d
```

Run kafka docker images:
```
sudo docker compose up kafka-ui kafka-1 -d
```

Set application config path:
```
export CONFIG_PATH=./configs/appsettings.json
```

Run the application:
```
go run ./cmd/app/main.go
```

### How it works from kafka point of view

Kafka producer
```
+--------------+                                      +-------------+
| Auth Service | <- Consume   [sms code]   <- Produce | Sms Service |
+--------------+                                      +-------------+
```

Kafka consumer
```
+--------------+                                      +-------------+
| Auth Service | Produce -> [phone number] Consume -> | Sms Service |
+--------------+                                      +-------------+
```