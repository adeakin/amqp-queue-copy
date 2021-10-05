# amqp-queue-copy
Copies amqp queues out to json files. Loads them back onto exchanges. Useful for replicating queues for testing/debugging purposes

## Building
```
go build
```

## Usage
```
amqp-queue-copy help
amqp-queue-copy load -help
amqp-queue-copy copy -help
````
````
NAME:
   amqp-queue-copy - Consumes messages from an AMQP queue into a JSON file / loads that file onto an exchange

USAGE:
   amqp-queue-copy [global options] command [command options] [arguments...]

COMMANDS:
   Copy, copy, c  Copy contents of queue to file
   Load, load, l  Copy contents of file to into queue/exchange
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
````
## Example
````
amqp-queue-copy copy --URI=amqp://rabbit:rabbit@localhost:5672/vhost --queue=test_queue --max=100
````
````json
[
    {
		"RoutingKey": "RoutingKey",
		"ContentType": "text/plain",
		"Exchange": "ExchangeName",
		"Body": "T2JqAQQUYXZyby5jb2RlYwhudWxsFmF2cm8uc2NoZW1h5Bx7InR5cGUiOiAicmVjb3JkIiwgIm5hbWUiOiAiQ29EIiwgIm5hbWVzcGFjZSI6ICJjb20uZXhhY3RlYXJ0aC5jb2QuZ2VuZXJhdGVkIiwgImZpZWxkcyI6IFt7Im5hbWUiOiAidXVpZCIsICJ0eXBlIjogInN0cmluZyJ9LCB7Im5hbWUiOiAicGF5bG9hZF9zY2hlbWEiLCAidHlwZSI6IHsidHlw"
    },
    ...
]
`````