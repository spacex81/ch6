package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var ctx context.Context

type Request struct {
	URL string `json:"url"`
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

	return feed.Entries, nil
}

func ParserHandler(c *gin.Context) {
	var request Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	entries, err := GetFeedEntries(request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while parsing the rss feed"})
		return
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	for _, entry := range entries[2:] {
		collection.InsertOne(ctx, bson.M{
			"title":     entry.Title,
			"thumbnail": entry.Thumbnail.URL,
			"url":       entry.Link.Href,
		})
	}

	c.JSON(http.StatusOK, entries)
}

func init() {
	ctx = context.Background()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
}

func main() {
	router := gin.Default()
	router.POST("/parse", ParserHandler)
	router.Run(":5000")
}
