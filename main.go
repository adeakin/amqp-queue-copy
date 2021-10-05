package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/streadway/amqp"
	"github.com/urfave/cli/v2"
)

type AMQPMessage struct {
	RoutingKey  string `json:"RoutingKey"`
	ContentType string `json:"ContentType"`
	Exchange    string `json:"Exchange"`
	Body        []byte `json:"Body"`
}

func main() {
	app := &cli.App{
		Name:  "amqp-queue-copy",
		Usage: "Consumes messages from an AMQP queue into a JSON file / loads that file onto an exchange",
		Commands: []*cli.Command{
			{
				Name:    "Copy",
				Aliases: []string{"copy", "c"},
				Usage:   "Copy contents of queue to file",
				Action:  copyQueue,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "URI",
						Aliases: []string{"uri"},
						Usage:   "AMQP URI e.g. amqp://user:pass@host:port/vhost",
					},
					&cli.StringFlag{
						Name:    "Queue",
						Aliases: []string{"queue", "q"},
						Usage:   "Queue to copy out to file ",
					},
					&cli.StringFlag{
						Name:    "File",
						Aliases: []string{"file", "f"},
						Usage:   "File to copy contents of queue into",
						Value:   "queue.json",
					},
					&cli.IntFlag{
						Name:    "Max",
						Aliases: []string{"max", "m"},
						Usage:   "Stop after this many messages",
					},
				},
			},
			{
				Name:    "Load",
				Aliases: []string{"load", "l"},
				Usage:   "Copy contents of file to into queue/exchange",
				Action:  loadFile,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "URI",
						Aliases: []string{"uri"},
						Usage:   "AMQP URI e.g. amqp://user:pass@host:port/vhost",
					},
					&cli.StringFlag{
						Name:    "Exchange",
						Aliases: []string{"exchange", "e"},
						Usage:   "Exchange to load file into",
					},
					&cli.StringFlag{
						Name:    "File",
						Aliases: []string{"file", "f"},
						Usage:   "File to read into queue",
						Value:   "queue.json",
					},
					&cli.StringFlag{
						Name:    "RoutingKey",
						Aliases: []string{"routingkey", "r"},
						Usage:   "Overwrite the routing key used with this value",
					},
					&cli.IntFlag{
						Name:    "Max",
						Aliases: []string{"max", "m"},
						Usage:   "Stop after this many messages",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func connectToRabbit(uri string) (conn *amqp.Connection, ch *amqp.Channel, err error) {
	var uri_valid bool
	if uri_valid, err = regexp.MatchString("amqp(s)*://(.+):(.+)@(.+):(\\d+)/(.+)", uri); !uri_valid || (err != nil) {
		return nil, nil, errors.New("rabbit URI invalid")
	}

	conn, err = amqp.Dial(uri)
	if err != nil {
		//Couldnt connect to Rabbit
		return nil, nil, err
	}

	ch, err = conn.Channel()
	if err != nil {
		//Couldnt open channel
		return nil, nil, err
	}
	return conn, ch, err
}

func loadFile(c *cli.Context) error {

	//load messages from json file
	file, err := ioutil.ReadFile(c.String("File"))
	if err != nil {
		//Problem opening file
		return err
	}

	messages := []AMQPMessage{}
	err = json.Unmarshal([]byte(file), &messages)
	if err != nil {
		//problem parsing file
		return err
	}

	//connect to rabbit
	conn, ch, err := connectToRabbit(c.String("URI"))
	if err != nil {
		return err
	}
	defer ch.Close()
	defer conn.Close()
	fmt.Printf("Connected to %s \n", c.String("URI"))

	exchange := c.String("Exchange")
	routingkey := c.String("RoutingKey")
	max := c.Int("Max")
	for i := range messages {
		//override exchange from file with cli flag
		if len(exchange) > 0 {
			messages[i].Exchange = exchange
		}
		//override routingkey from file with cli flag
		if len(routingkey) > 0 {
			messages[i].RoutingKey = routingkey
		}

		err = ch.Publish(
			messages[i].Exchange,
			messages[i].RoutingKey,
			false,
			false,
			amqp.Publishing{
				ContentType: messages[i].ContentType,
				Body:        []byte(messages[i].Body),
			},
		)
		if err != nil {
			//error writing message to exchange
			return err
		}

		if i > max {
			//stop after writing certain number of messages
			break
		}
		fmt.Printf("\rWrote %d messages", i+1)
	}
	fmt.Println("\nFinished")

	return nil
}

func copyQueue(c *cli.Context) error {
	//connect to rabbit
	conn, ch, err := connectToRabbit(c.String("URI"))
	if err != nil {
		return err
	}
	defer ch.Close()
	defer conn.Close()
	fmt.Printf("Connected to %s \n", c.String("URI"))

	max := c.Int("Max")
	cur_count := 0
	var parsed_msgs []AMQPMessage
	//collect messages from queue
	for {
		cur_count += 1
		msg, ok, err := ch.Get(c.String("Queue"), false)
		if err != nil {
			return err
		}
		if !ok {
			//stop when queue is empty
			fmt.Print("Queue empty")
			break
		}

		parsed := AMQPMessage{msg.RoutingKey, msg.ContentType, msg.Exchange, msg.Body}
		parsed_msgs = append(parsed_msgs, parsed)
		msg.Ack(false)
		fmt.Printf("\rCollected %d messages", cur_count)
		if max > 0 && (cur_count >= max) {
			//stop after collecting certain number of messages
			break
		}
	}
	fmt.Println()

	//write out queue contents to file
	if len(parsed_msgs) > 0 {
		filename := c.String("File")
		json, err := json.MarshalIndent(parsed_msgs, "", "\t")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filename, json, 0644)
		if err != nil {
			return err
		}
		fmt.Printf("Wrote messages to %s \n", filename)
	}

	return nil
}
