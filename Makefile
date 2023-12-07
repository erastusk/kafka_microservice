producer:
	@go build -o bin/producer ./producer
	@./bin/producer

receiver:
	@go build -o bin/receiver ./data_receiver_kafka_producer
	@./bin/receiver

calculator:
	@go build -o bin/calculator ./kafka_reader
	@./bin/calculator


.PHONY: producer docker receiver calculator
