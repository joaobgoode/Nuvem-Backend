package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
)

type ProductNoID struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Product struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

var (
	supabaseUrl string
	supabaseKey string
	port        string
	supabase    *supa.Client
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	supabaseUrl = os.Getenv("SUPA_URL")
	supabaseKey = os.Getenv("SUPA_KEY")
	supabase = supa.CreateClient(supabaseUrl, supabaseKey)

	port = os.Getenv("PORT")
}

func AllProductsHandler(w http.ResponseWriter, r *http.Request) {
	var results []Product
	err := supabase.DB.From("products").Select("*").Execute(&results)
	if err != nil {
		w.WriteHeader(400)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
	}
}

func NewProductHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	descriptionStr := r.PathValue("description")
	description := descriptionStr[2:] // Remove "d=" from description
	priceStr := r.PathValue("price")
	price, _ := strconv.ParseFloat(priceStr, 64)

	p := ProductNoID{name, description, price}

	var product []Product
	err := supabase.DB.From("products").Insert(p).Execute(&product)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
	}
}

func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var products []Product
	err := supabase.DB.From("products").Delete().Eq("id", id).Execute(&products)
	if err != nil {
		panic(err)
	}

	log.Printf("Deleted: %v\n", products)
	w.WriteHeader(200)
}

func EditProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	name := r.PathValue("name")
	descriptionStr := r.PathValue("description")

	var description string
	_, err := fmt.Sscanf(descriptionStr, "d=%s", &description)
	if err != nil {
		panic(err)
	}

	priceStr := r.PathValue("price")
	price, _ := strconv.ParseFloat(priceStr, 64)

	p := Product{id, name, description, price}

	var results []Product
	err = supabase.DB.From("products").Update(p).Eq("id", idStr).Execute(&results)
	if err != nil {
		panic(err)
	}

	log.Printf("Updated: %v\n", results)
	w.WriteHeader(200)
}

func main() {
	http.HandleFunc("GET /products/", AllProductsHandler)
	http.HandleFunc("POST /new/{name}/{description}/{price}", NewProductHandler)
	http.HandleFunc("DELETE /delete/{id}", DeleteProductHandler)
	http.HandleFunc("PUT /edit/{id}/{name}/{description}/{price}", EditProductHandler)

	http.ListenAndServe(port, nil)
}
