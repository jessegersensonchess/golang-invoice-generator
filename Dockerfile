FROM golang:1.17
WORKDIR "/app"
COPY . .

RUN CGO_ENABLED=0 go build -o invoice &&\
	chmod +x invoice && \
	cp invoice /usr/local/bin/ && \
	echo "Example usage: ./invoice -k 10.00 generate /tmp/test.csv"

