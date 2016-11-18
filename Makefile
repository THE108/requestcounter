build: deps test
	go build -o requestcounter main.go

test:
	go test `glide nv`

deps:
	glide install
