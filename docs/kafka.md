### create kafka topic
```sh
docker compose -f ./docker-compose.yml exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --create --topic gopan.transcode.tasks --partitions 1 --replication-factor 1 2>&1
```