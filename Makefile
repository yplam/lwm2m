.PHONY: run_simple run_mqtt

run_simple:
	go run examples/simple/*.go

run_mqtt:
	go run examples/lwm2mqtt/*.go
