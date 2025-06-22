package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bxcodec/faker/v4"
	"github.com/go-resty/resty/v2"
	"github.com/valyala/fasttemplate"
)

type APIClient struct {
	client *resty.Client
	token  string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		client: resty.New().SetBaseURL(baseURL),
	}
}

// Генерация тестовых данных
func generateUserData() map[string]interface{} {
	return map[string]interface{}{
		"username": faker.Username(),
		"email":    faker.Email(),
		"password": faker.Password(),
	}
}

func generateWishlistData(userID string) map[string]interface{} {
	return map[string]interface{}{
		"title":       faker.Sentence(),
		"description": faker.Paragraph(),
		"user_id":     userID,
	}
}

func generateItemData(wishlistID string) map[string]interface{} {
	products := []string{"iPhone", "MacBook", "Watch", "iPad", "AirPods"}
	return map[string]interface{}{
		"name":        fmt.Sprintf("%s %s", products[rand.Intn(len(products))], faker.Word()),
		"description": faker.Sentence(),
		"price":       fmt.Sprintf("%.2f", rand.Float64()*1000+100),
		"link":        faker.URL(),
		"wishlist_id": wishlistID,
	}
}

// API методы
func (c *APIClient) Register(userData map[string]interface{}) (string, error) {
	resp, err := c.client.R().
		SetBody(userData).
		Post("/auth/register")

	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (c *APIClient) Login(username, password string) error {
	resp, err := c.client.R().
		SetBody(map[string]string{
			"username": username,
			"password": password,
		}).
		Post("/auth/login")

	if err != nil {
		return err
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return err
	}

	c.token = result.Token
	c.client.SetAuthToken(result.Token)
	return nil
}

func (c *APIClient) CreateWishlist(data map[string]interface{}) (string, error) {
	resp, err := c.client.R().
		SetBody(data).
		SetAuthToken(c.token).
		Post("/api/wishlists")

	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (c *APIClient) AddWishlistItem(wishlistID string, data map[string]interface{}) (string, error) {
	resp, err := c.client.R().
		SetBody(data).
		SetAuthToken(c.token).
		Post(fmt.Sprintf("/api/wishlists/%s/items", wishlistID))

	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func main() {
	// Инициализация генератора случайных цифр
	rand.NewSource(time.Now().UnixNano())

	// 1. Инициализация клиента
	api := NewAPIClient("http://localhost:8080")

	// 2. Регистрация и авторизация пользователя
	userData := generateUserData()
	userID, err := api.Register(userData)
	if err != nil {
		log.Fatalf("Registration failed: %v", err)
	}
	fmt.Printf("Registered user: %s (%s)\n", userData["username"], userID)

	if err := api.Login(userData["username"].(string), userData["password"].(string)); err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Println("Successfully logged in")

	// 3. Создание списка желаний
	wishlistData := generateWishlistData(userID)
	wishlistID, err := api.CreateWishlist(wishlistData)
	if err != nil {
		log.Fatalf("Wishlist creation failed: %v", err)
	}
	fmt.Printf("Created wishlist: %s (%s)\n", wishlistData["title"], wishlistID)

	// 4. Добавление нескольких элементов
	for i := 0; i < 3; i++ {
		itemData := generateItemData(wishlistID)
		itemID, err := api.AddWishlistItem(wishlistID, itemData)
		if err != nil {
			log.Printf("Failed to add item: %v", err)
			continue
		}
		fmt.Printf("Added item: %s (%.2f) -> %s\n",
			itemData["name"], itemData["price"], itemID)
	}

	template := `User {{username}} ({{email}}) created wishlist "{{title}}" with items: {{items}}`
	t := fasttemplate.New(template, "{{", "}}")

	itemsList := []string{
		fmt.Sprintf("%s (%.2f)", generateItemData(wishlistID)["name"], generateItemData(wishlistID)["price"]),
		fmt.Sprintf("%s (%.2f)", generateItemData(wishlistID)["name"], generateItemData(wishlistID)["price"]),
	}

	result := t.ExecuteString(map[string]interface{}{
		"username": userData["username"],
		"email":    userData["email"],
		"title":    wishlistData["title"],
		"items":    strings.Join(itemsList, ", "),
	})

	fmt.Println("\n=== Added Wishlist item ===")
	fmt.Println(result)
}
