package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client

type SubmitRequest struct {
	QuesID string `json:"quesId"`
	Language string `json:"language"`
	Code   string `json:"code"`
}
type Question struct {
	QuesID    string     `bson:"quesId"`
	TestCases []TestCase `bson:"testCases"`
}

func getTestCases(client *mongo.Client, quesId string) ([]TestCase, error) {
	collection := client.Database("judge").Collection("questions")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var question Question
	err :=collection.FindOne(ctx, bson.M{"quesId": quesId}).Decode(&question)
	if err != nil {
		return nil, err
	}

	return question.TestCases, nil
}

type TestResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Output   string `json:"output"`
	Passed   bool   `json:"passed"`
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	var req SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
if req.Language != "go" {
		http.Error(w, "Only Go language is supported", http.StatusBadRequest)
		return
	}
	testCases, err := getTestCases(mongoClient, req.QuesID)
	if err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	results := RunAllTestCases(req.Code, testCases)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
func connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Set a timeout for the connection attempt
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, ctx, cancel, err
	}

	// Ping the primary to verify connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, ctx, cancel, err
	}

	return client, ctx, cancel, nil
}
func closeMongo(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}
func main() {
	http.HandleFunc("/submit", submitHandler)

	uri := "mongodb://localhost:27017"

	client, ctx, cancel, err := connect(uri)
	mongoClient = client

	if err != nil {
		log.Fatal(err)
	}

	// Defer the Disconnect call to ensure the connection is closed when the main function returns
	defer closeMongo(client, ctx, cancel)

	fmt.Println("Connected to MongoDB!")
	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
