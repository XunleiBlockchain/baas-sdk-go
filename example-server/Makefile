BUILD_TAGS?=

default: clean build

clean:
	rm -rf sdk-server

build:
	go build -tags '$(BUILD_TAGS)' -o sdk-server .
