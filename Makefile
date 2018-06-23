run:
	go run *.go

install:
	go install
	sudo mv ~/go/bin/forwarder /usr/local/bin/

dep-install:
	glide install
