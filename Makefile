.PHONY: proto build dev

proto:
	export PATH=$(HOME)/go/bin:$$PATH && protoc --go_out=. --go_opt=paths=source_relative backend/p2p/proto/redware.proto

build:
	wails build

dev:
	wails dev
