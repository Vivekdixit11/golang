package main

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

// CORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handler function
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fmt.Fprintf(w, "Server is up and connected to MongoDB")
	case "/login":
		loginHandler(w, r)
	case "/create-course":
		createCourseHandler(w, r)
	case "/courses":
		getCoursesHandler(w, r)
	case "/courses/upcoming":
		getUpcomingCoursesHandler(w, r)
	case "/courses/active":
		getActiveCoursesHandler(w, r)
	default:
		http.Error(w, "Bhai API galat call kar rha hai ek baar dekh le", http.StatusNotFound)
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

type Module struct {
	Title string `json:"title" bson:"title"`
	PDF   string `json:"pdf" bson:"pdf"`
}

type Course struct {
	Title       string   `json:"title" bson:"title"`
	Description string   `json:"description" bson:"description"`
	Duration    string   `json:"duration" bson:"duration"`
	Modules     []Module `json:"modules" bson:"modules"`
	Type        string   `json:"type" bson:"type"`
}

func createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var course Course
	err := json.NewDecoder(r.Body).Decode(&course)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !validateCourse(course) {
		http.Error(w, "Invalid course data", http.StatusBadRequest)
		return
	}

	collection := client.Database("lms").Collection("courses")
	_, err = collection.InsertOne(context.TODO(), course)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Course created successfully"})
}

func validateCourse(course Course) bool {
	if len(course.Title) > 20 || len(course.Description) > 150 {
		return false
	}
	for _, module := range course.Modules {
		if len(module.Title) > 20 {
			return false
		}
	}
	return true
}

func getCoursesHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("lms").Collection("courses")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var courses []Course
	if err = cursor.All(context.TODO(), &courses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func getUpcomingCoursesHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("lms").Collection("courses")
	cursor, err := collection.Find(context.TODO(), bson.M{"type": "upcoming"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var courses []Course
	if err = cursor.All(context.TODO(), &courses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func getActiveCoursesHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("lms").Collection("courses")
	cursor, err := collection.Find(context.TODO(), bson.M{"type": "Active"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var courses []Course
	if err = cursor.All(context.TODO(), &courses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the retrieved courses for debugging
	fmt.Printf("Retrieved active courses: %+v\n", courses)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/create-course", createCourseHandler)
	mux.HandleFunc("/courses", getCoursesHandler)
	mux.HandleFunc("/courses/upcoming", getUpcomingCoursesHandler)
	mux.HandleFunc("/courses/active", getActiveCoursesHandler)

	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
