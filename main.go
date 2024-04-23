package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WordsStruct struct {
	Date  string
	Words []string
}


func init() {
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func main() {
	
	router := mux.NewRouter()
	var st string

	router.HandleFunc("/words", func(w http.ResponseWriter, r *http.Request) {

		collection := connectMangoDB()

		st = fetchTodaysRecord(collection)

		if st == "fetchedfreshrecord" {
			st = fetchTodaysRecord(collection)
		}

		fmt.Fprint(w, st)
	})

	http.ListenAndServe(":8000", router)

}

func fetchWords() []string {
	words := []string{}
	for i := 5; i < 11; i++ {
		baseURL := "https://random-word-api.herokuapp.com/word?length=" + strconv.Itoa(i)
		response, err := http.Get(baseURL)

		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		word := string(responseData[2 : len(responseData)-2])
		// fmt.Println(word)

		words = append(words, word)

	}

	return words

}

func connectMangoDB() *mongo.Collection {

	uri := os.Getenv("URI")
    if uri == "" {
        log.Fatal("URI not set")
    }

	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("unscrabble").Collection("unsc")

	return collection
}

func fetchTodaysRecord(collection *mongo.Collection) string {
	var final string
	var timeformat = time.Now().Format("01-02-2006")

	filter := bson.D{{Key: "date", Value: timeformat}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	var results []WordsStruct
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	if len(results) == 0 {
		//No records are present thus creating todays words and inserting to the db
		fmt.Println("No records present,fetching todays record!")
		words := fetchWords()
		row1 := WordsStruct{Date: timeformat, Words: words}
		_, err = collection.InsertOne(context.TODO(), row1)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		return "fetchedfreshrecord"
	} else {
		for _, result := range results {
			res, _ := bson.MarshalExtJSON(result, false, false)
			// Unmarshal BSON byte array into a map
			var data map[string]interface{}
			err := json.Unmarshal(res, &data)
			if err != nil {
				fmt.Println("Error unmarshalling BSON data:", err)
			}
			// fmt.Println(string(res))
			final = string(res)
		}

		return final
	}

}
