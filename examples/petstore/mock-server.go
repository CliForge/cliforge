package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed petstore-api.yaml
var specFile embed.FS

// Data models
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Pet struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  Category  `json:"category"`
	Status    string    `json:"status"`
	PhotoUrls []string  `json:"photoUrls,omitempty"`
	Tags      []Tag     `json:"tags,omitempty"`
	Age       int       `json:"age,omitempty"`
	Price     float64   `json:"price,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Store struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Capacity int    `json:"capacity"`
}

type Order struct {
	ID       int64     `json:"id"`
	PetID    int64     `json:"petId"`
	Quantity int       `json:"quantity"`
	ShipDate time.Time `json:"shipDate,omitempty"`
	Status   string    `json:"status"`
	Complete bool      `json:"complete"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone,omitempty"`
}

type Context struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Active  bool   `json:"active"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Mock database
type MockDB struct {
	mu     sync.RWMutex
	pets   map[int64]*Pet
	stores map[int64]*Store
	orders map[int64]*Order
	users  map[int64]*User
	nextID int64
}

var db = &MockDB{
	pets:   make(map[int64]*Pet),
	stores: make(map[int64]*Store),
	orders: make(map[int64]*Order),
	users:  make(map[int64]*User),
	nextID: 1,
}

func init() {
	// Initialize with sample data
	now := time.Now()

	// Sample pets
	db.pets[1] = &Pet{
		ID:   1,
		Name: "Buddy",
		Category: Category{
			ID:   1,
			Name: "dog",
		},
		Status:    "available",
		Age:       3,
		Price:     250.00,
		CreatedAt: now,
		UpdatedAt: now,
		Tags: []Tag{
			{ID: 1, Name: "friendly"},
			{ID: 2, Name: "trained"},
		},
	}

	db.pets[2] = &Pet{
		ID:   2,
		Name: "Whiskers",
		Category: Category{
			ID:   2,
			Name: "cat",
		},
		Status:    "available",
		Age:       2,
		Price:     150.00,
		CreatedAt: now,
		UpdatedAt: now,
		Tags: []Tag{
			{ID: 3, Name: "calm"},
		},
	}

	db.pets[3] = &Pet{
		ID:   3,
		Name: "Tweety",
		Category: Category{
			ID:   3,
			Name: "bird",
		},
		Status:    "pending",
		Age:       1,
		Price:     75.00,
		CreatedAt: now,
		UpdatedAt: now,
	}

	db.pets[4] = &Pet{
		ID:   4,
		Name: "Nemo",
		Category: Category{
			ID:   4,
			Name: "fish",
		},
		Status:    "available",
		Age:       1,
		Price:     25.00,
		CreatedAt: now,
		UpdatedAt: now,
	}

	db.pets[5] = &Pet{
		ID:   5,
		Name: "Rex",
		Category: Category{
			ID:   5,
			Name: "reptile",
		},
		Status:    "sold",
		Age:       5,
		Price:     500.00,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Sample stores
	db.stores[1] = &Store{
		ID:       1,
		Name:     "Downtown Petstore",
		Location: "123 Main St, New York, NY",
		Capacity: 100,
	}

	db.stores[2] = &Store{
		ID:       2,
		Name:     "Westside Pets",
		Location: "456 Oak Ave, Los Angeles, CA",
		Capacity: 75,
	}

	// Sample users
	db.users[1] = &User{
		ID:        1,
		Username:  "johndoe",
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1-555-0100",
	}

	db.users[2] = &User{
		ID:        2,
		Username:  "janedoe",
		Email:     "jane@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Phone:     "+1-555-0101",
	}

	// Sample orders
	db.orders[1] = &Order{
		ID:       1,
		PetID:    3,
		Quantity: 1,
		Status:   "placed",
		Complete: false,
	}

	db.nextID = 100
}

func (d *MockDB) getNextID() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	id := d.nextID
	d.nextID++
	return id
}

// Handlers
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	data, err := specFile.ReadFile("petstore-api.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	_, _ = w.Write(data)
}

func listPets(w http.ResponseWriter, r *http.Request) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = val
		}
	}

	var pets []*Pet
	for _, pet := range db.pets {
		if status != "" && pet.Status != status {
			continue
		}
		if category != "" && pet.Category.Name != category {
			continue
		}
		pets = append(pets, pet)
	}

	// Apply pagination
	start := offset
	if start > len(pets) {
		start = len(pets)
	}
	end := start + limit
	if end > len(pets) {
		end = len(pets)
	}

	result := pets[start:end]

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func createPet(w http.ResponseWriter, r *http.Request) {
	var pet Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validation
	if pet.Name == "" {
		sendError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	if pet.Category.Name == "" {
		sendError(w, http.StatusBadRequest, "Category is required", nil)
		return
	}

	db.mu.Lock()
	pet.ID = db.getNextID()
	pet.CreatedAt = time.Now()
	pet.UpdatedAt = pet.CreatedAt
	if pet.Status == "" {
		pet.Status = "available"
	}
	db.pets[pet.ID] = &pet
	db.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pet)
}

func getPet(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/pets/")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid pet ID", err.Error())
		return
	}

	db.mu.RLock()
	pet, exists := db.pets[id]
	db.mu.RUnlock()

	if !exists {
		sendError(w, http.StatusNotFound, "Pet not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pet)
}

func updatePet(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/pets/")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid pet ID", err.Error())
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	pet, exists := db.pets[id]
	if !exists {
		sendError(w, http.StatusNotFound, "Pet not found", nil)
		return
	}

	var updates Pet
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Update fields
	if updates.Name != "" {
		pet.Name = updates.Name
	}
	if updates.Status != "" {
		pet.Status = updates.Status
	}
	if updates.Price > 0 {
		pet.Price = updates.Price
	}
	pet.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pet)
}

func deletePet(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/pets/")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid pet ID", err.Error())
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.pets[id]; !exists {
		sendError(w, http.StatusNotFound, "Pet not found", nil)
		return
	}

	delete(db.pets, id)
	w.WriteHeader(http.StatusNoContent)
}

func streamPetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/pets/")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid pet ID", err.Error())
		return
	}

	db.mu.RLock()
	pet, exists := db.pets[id]
	db.mu.RUnlock()

	if !exists {
		sendError(w, http.StatusNotFound, "Pet not found", nil)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial event
	_, _ = fmt.Fprintf(w, "event: status-change\n")
	_, _ = fmt.Fprintf(w, "data: {\"timestamp\":\"%s\",\"event\":\"status-change\",\"data\":{\"message\":\"Current status: %s\",\"petId\":%d}}\n\n", time.Now().Format(time.RFC3339), pet.Status, pet.ID)
	flusher.Flush()

	// Simulate periodic updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	count := 0
	events := []string{"status-change", "location-update", "health-check"}
	messages := []string{
		"Pet is doing well",
		"Feeding completed",
		"Exercise time",
		"Health check performed",
		"Location updated",
	}

	for {
		select {
		case <-ticker.C:
			event := events[count%len(events)]
			message := messages[count%len(messages)]

			_, _ = fmt.Fprintf(w, "event: %s\n", event)
			_, _ = fmt.Fprintf(w, "data: {\"timestamp\":\"%s\",\"event\":\"%s\",\"data\":{\"message\"%s\",\"petId\":%d}}\n\n",
				time.Now().Format(time.RFC3339), event, message, pet.ID)
			flusher.Flush()

			count++
			if count >= 10 {
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}

func listStores(w http.ResponseWriter, r *http.Request) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var stores []*Store
	for _, store := range db.stores {
		stores = append(stores, store)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stores)
}

func getStoreCapacity(w http.ResponseWriter, r *http.Request) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	total := 0
	used := 0
	for _, store := range db.stores {
		total += store.Capacity
	}
	used = len(db.pets)
	available := total - used

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total":     total,
		"used":      used,
		"available": available,
	})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	db.mu.Lock()
	order.ID = db.getNextID()
	order.Status = "placed"
	order.Complete = false
	db.orders[order.ID] = &order
	db.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)

	// Simulate async status updates
	go func() {
		time.Sleep(10 * time.Second)
		db.mu.Lock()
		if o, exists := db.orders[order.ID]; exists {
			o.Status = "approved"
		}
		db.mu.Unlock()

		time.Sleep(20 * time.Second)
		db.mu.Lock()
		if o, exists := db.orders[order.ID]; exists {
			o.Status = "delivered"
			o.Complete = true
		}
		db.mu.Unlock()
	}()
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/orders/")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	db.mu.RLock()
	order, exists := db.orders[id]
	db.mu.RUnlock()

	if !exists {
		sendError(w, http.StatusNotFound, "Order not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var users []*User
	for _, user := range db.users {
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/users/")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	db.mu.RLock()
	user, exists := db.users[id]
	db.mu.RUnlock()

	if !exists {
		sendError(w, http.StatusNotFound, "User not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func listContexts(w http.ResponseWriter, r *http.Request) {
	contexts := []Context{
		{Name: "development", BaseURL: "http://localhost:8080", Active: true},
		{Name: "staging", BaseURL: "https://staging-api.petstore.example.com", Active: false},
		{Name: "production", BaseURL: "https://api.petstore.example.com", Active: false},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(contexts)
}

func listRegions(w http.ResponseWriter, r *http.Request) {
	regions := []map[string]string{
		{"id": "us-east-1", "display_name": "US East (N. Virginia)"},
		{"id": "us-west-2", "display_name": "US West (Oregon)"},
		{"id": "eu-west-1", "display_name": "Europe (Ireland)"},
		{"id": "ap-southeast-1", "display_name": "Asia Pacific (Singapore)"},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(regions)
}

func verifyCredentials(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"valid": true})
}

func verifyQuotas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"sufficient": true})
}

func getDeploymentReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ready": true})
}

func createDeployment(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)

	deployment := map[string]interface{}{
		"id":      db.getNextID(),
		"app_id":  body["app_id"],
		"version": body["version"],
		"status":  "deploying",
		"url":     "https://app.example.com",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(deployment)
}

func getDeployment(w http.ResponseWriter, r *http.Request) {
	id, _ := extractID(r.URL.Path, "/deployments/")

	deployment := map[string]interface{}{
		"id":     id,
		"status": "deployed",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(deployment)
}

// Helper functions
func extractID(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	// Remove any trailing path segments
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func sendError(w http.ResponseWriter, code int, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(Error{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func main() {
	// OpenAPI spec
	http.HandleFunc("/openapi.yaml", corsMiddleware(serveOpenAPISpec))

	// Pets
	http.HandleFunc("/pets", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			listPets(w, r)
		case "POST":
			createPet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/pets/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/status/stream") {
			streamPetStatus(w, r)
			return
		}

		switch r.Method {
		case "GET":
			getPet(w, r)
		case "PUT":
			updatePet(w, r)
		case "DELETE":
			deletePet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Stores
	http.HandleFunc("/stores", corsMiddleware(listStores))
	http.HandleFunc("/stores/capacity", corsMiddleware(getStoreCapacity))

	// Orders
	http.HandleFunc("/orders", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			createOrder(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/orders/", corsMiddleware(getOrder))

	// Users
	http.HandleFunc("/users", corsMiddleware(listUsers))
	http.HandleFunc("/users/", corsMiddleware(getUser))

	// Contexts
	http.HandleFunc("/contexts", corsMiddleware(listContexts))

	// Regions
	http.HandleFunc("/regions", corsMiddleware(listRegions))

	// Internal
	http.HandleFunc("/credentials/verify", corsMiddleware(verifyCredentials))
	http.HandleFunc("/quotas/verify", corsMiddleware(verifyQuotas))

	// Deployments
	http.HandleFunc("/deployments/readiness", corsMiddleware(getDeploymentReadiness))
	http.HandleFunc("/deployments", corsMiddleware(createDeployment))
	http.HandleFunc("/deployments/", corsMiddleware(getDeployment))

	port := ":8080"
	log.Printf("Starting Petstore Mock API server on %s", port)
	log.Printf("OpenAPI spec: http://localhost%s/openapi.yaml", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
