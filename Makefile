.PHONY: clean

twitter.out: twitter.go oauth.go utils.go
	go build -o $@ $^

clean:
	rm -f *.out
