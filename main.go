package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
	clientOptions := options.Client().ApplyURI("mongodb+srv://vivekdixit48313:A800900plmA@cluster0.nydox.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
}

// Handler function
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fmt.Fprintf(w, "Server is up and connected to MongoDB")
	case "/login":
		loginHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		LoginID  string `json:"login_id"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := client.Database("lms").Collection("admin_login")
	var result struct {
		LoginID  string `bson:"login_id"`
		Password string `bson:"password"`
	}

	err = collection.FindOne(context.TODO(), bson.M{"login_id": credentials.LoginID, "password": credentials.Password}).Decode(&result)
	response := map[string]bool{"success": err == nil}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
