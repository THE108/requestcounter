build: deps test
	go build -o requestcount main.go

test:
	go test `glide nv`

deps:
	glide install
