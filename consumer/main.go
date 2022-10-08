package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Request struct {
	URL string `xml:"json:"url"`
}

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`
	Title string `xml:"title"`
}

func GetFeedEntries(url string) ([]Entry, error) {
	cmd := exec.Command("curl", "-A", "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/81.0", "-O", url)

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	xmlFile, err := os.Open(".rss")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened .rss")
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var feed Feed
	xml.Unmarshal(byteValue, &feed)

	fmt.Print("feed")
	fmt.Print(feed)

	return feed.Entries, nil
}

// RABBITMQ_URI="amqp://user:password@localhost:5672/" RABBITMQ_QUEUE=rss_urls MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin&readPreference=primary&appname=MongoDB%20Compass&ssl=false" MONGO_DATABASE=demo go run main.go
func main() {

	// mongodb
	ctx := context.Background()
	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	defer mongoClient.Disconnect(ctx)

	// mqrabbit
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}
	defer amqpConnection.Close()

	channelAmqp, _ := amqpConnection.Channel()
	defer channelAmqp.Close()

	forever := make(chan bool)

	msgs, err := channelAmqp.Consume(
		os.Getenv("RABBITMQ_QUEUE"),
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var request Request
			json.Unmarshal(d.Body, &request)

			log.Println("RSS URL:", request.URL)

			entries, _ := GetFeedEntries(request.URL)

			// fmt.Println(entries)

			collection := mongoClient.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
			fmt.Println(len(entries))
			for _, entry := range entries {
				collection.InsertOne(ctx, bson.M{
					"title":     entry.Title,
					"thumbnail": entry.Thumbnail.URL,
					"url":       entry.Link.Href,
				})
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
