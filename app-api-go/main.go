package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const mongo_db = "langdb"
const mongo_collection = "languages"
const mongo_default_conn_str = "mongodb://mongo-0.mongo,mongo-1.mongo,mongo-2.mongo:27017/langdb"
const mongo_default_username = "admin"
const mongo_default_password = "password"

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

var c *mongo.Client

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func getClient() *mongo.Client {
	mongoconnstr := getEnv("MONGO_CONN_STR", mongo_default_conn_str)
	mongousername := getEnv("MONGO_USERNAME", mongo_default_username)
	// mongopassword := getEnv("MONGO_PASSWORD", mongo_default_password)

	fmt.Println("MongoDB connection details:")
	fmt.Println("MONGO_CONN_STR:" + mongoconnstr)
	fmt.Println("MONGO_USERNAME:" + mongousername)
	fmt.Println("MONGO_PASSWORD:")
	fmt.Println("attempting mongodb backend connection...")

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://localhost:mongodb@cluster0.rnp1coz.mongodb.net/?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// clientOptions := options.Client().ApplyURI(mongoconnstr)

	//test if auth is enabled or expected,
	//for demo purposes when we setup mongo as a replica set using a StatefulSet resource in K8s auth is disabled

	// if clientOptions.Auth != nil {
	// 	clientOptions.Auth.Username = mongousername
	// 	clientOptions.Auth.Password = mongopassword
	// }

	// options.Client().SetMaxConnIdleTime(60000)
	// options.Client().SetHeartbeatInterval(5 * time.Second)

	// client, err := mongo.NewClient(clientOptions)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// err = client.Connect(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return client
}

func init() {
	c = getClient()
	err := c.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("couldn't connect to the database", err)
	} else {
		log.Println("connected!!")
	}
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop over the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.POST("/albums", postAlbums)
	router.GET("/albums/:id", getAlbumByID)

	router.Run("localhost:8080")
}
