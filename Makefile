mssim:
	go build -o mssim cmd/*.go	
clean:
	rm mssim

.PHONY: mssim

.DEFAULT_GOAL := all
all: mssim
