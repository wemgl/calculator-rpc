.PHONY: proto
proto:
	protoc --go_out=plugins=grpc:. calculatorpb/calculator.proto
