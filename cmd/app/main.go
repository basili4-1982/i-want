package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Структуры данных
type User struct {
	ID       string `json:"id"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Wishlist struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Item struct {
	ID          string `json:"id"`
	WishlistID  string `json:"wishlist_id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Link        string `json:"link"`
	IsPurchased bool   `json:"is_purchased"`
}

type SharedWishlist struct {
	ID         string `json:"id"`
	WishlistID string `json:"wishlist_id"`
	UserID     string `json:"user_id"`
	CanEdit    bool   `json:"can_edit"`
}

// In-memory хранилища
var (
	users           = make(map[string]User)
	wishlists       = make(map[string]Wishlist)
	items           = make(map[string]Item)
	sharedWishlists = make(map[string]SharedWishlist)
	mu              sync.RWMutex
)

func main() {
	r := gin.Default()

	// Группа маршрутов для аутентификации
	auth := r.Group("/auth")
	{
		auth.POST("/register", register)
		auth.POST("/login", login)
	}

	// Группа маршрутов для работы со списками желаний
	api := r.Group("/api", authMiddleware)
	{
		api.GET("/wishlists", getWishlists)
		api.POST("/wishlists", createWishlist)
		api.GET("/wishlists/:id", getWishlist)
		api.PUT("/wishlists/:id", updateWishlist)
		api.DELETE("/wishlists/:id", deleteWishlist)

		api.GET("/wishlists/:id/items", getItems)
		api.POST("/wishlists/:id/items", addItem)
		api.PUT("/wishlists/:id/items/:item_id", updateItem)
		api.DELETE("/wishlists/:id/items/:item_id", deleteItem)

		api.POST("/wishlists/:id/share", shareWishlist)
		api.GET("/shared", getSharedWishlists)
	}

	r.Run(":8080")
}

// Middleware для проверки аутентификации
func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// В реальном приложении здесь должна быть проверка JWT токена
	// Для упрощения просто проверяем, что пользователь существует
	mu.RLock()
	_, exists := users[token]
	mu.RUnlock()

	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Set("userID", token)
	c.Next()
}

// Хэлпер-функции
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Обработчики маршрутов
func register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Проверяем, существует ли пользователь
	for _, u := range users {
		if u.Username == user.Username || u.Email == user.Email {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
			return
		}
	}

	// Хэшируем пароль
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	// Создаем пользователя
	user.ID = uuid.New().String()
	user.Password = hashedPassword
	users[user.ID] = user

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	// Ищем пользователя
	var foundUser User
	for _, user := range users {
		if user.Username == credentials.Username {
			foundUser = user
			break
		}
	}

	if foundUser.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Проверяем пароль
	if !checkPasswordHash(credentials.Password, foundUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": foundUser.ID,
		"user": gin.H{
			"id":       foundUser.ID,
			"username": foundUser.Username,
			"email":    foundUser.Email,
		},
	})
}

func createWishlist(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var wishlist Wishlist
	if err := c.ShouldBindJSON(&wishlist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	wishlist.ID = uuid.New().String()
	wishlist.UserID = userID
	wishlist.CreatedAt = time.Now()
	wishlist.UpdatedAt = time.Now()

	wishlists[wishlist.ID] = wishlist

	c.JSON(http.StatusCreated, wishlist)
}

func getWishlists(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	mu.RLock()
	defer mu.RUnlock()

	var userWishlists []Wishlist
	for _, w := range wishlists {
		if w.UserID == userID {
			userWishlists = append(userWishlists, w)
		}
	}

	c.JSON(http.StatusOK, userWishlists)
}

func getWishlist(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")

	mu.RLock()
	defer mu.RUnlock()

	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	// Проверяем, что пользователь имеет доступ к списку
	if wishlist.UserID != userID && !hasSharedAccess(userID, wishlistID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, wishlist)
}

func updateWishlist(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")

	var update Wishlist
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	// Проверяем права на редактирование
	if wishlist.UserID != userID && !hasEditAccess(userID, wishlistID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Обновляем поля
	wishlist.Title = update.Title
	wishlist.Description = update.Description
	wishlist.UpdatedAt = time.Now()

	wishlists[wishlistID] = wishlist

	c.JSON(http.StatusOK, wishlist)
}

func deleteWishlist(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")

	mu.Lock()
	defer mu.Unlock()

	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	// Проверяем права на удаление (только владелец может удалить)
	if wishlist.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Удаляем список и связанные с ним элементы
	delete(wishlists, wishlistID)
	for itemID, item := range items {
		if item.WishlistID == wishlistID {
			delete(items, itemID)
		}
	}

	// Удаляем записи о совместном доступе
	for shareID, share := range sharedWishlists {
		if share.WishlistID == wishlistID {
			delete(sharedWishlists, shareID)
		}
	}

	c.Status(http.StatusNoContent)
}

func addItem(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")

	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Проверяем существование списка и права доступа
	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	if wishlist.UserID != userID && !hasEditAccess(userID, wishlistID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Создаем элемент
	item.ID = uuid.New().String()
	item.WishlistID = wishlistID
	item.IsPurchased = false

	items[item.ID] = item

	c.JSON(http.StatusCreated, item)
}

func getItems(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")

	mu.RLock()
	defer mu.RUnlock()

	// Проверяем существование списка и права доступа
	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	if wishlist.UserID != userID && !hasSharedAccess(userID, wishlistID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Собираем элементы списка
	var wishlistItems []Item
	for _, item := range items {
		if item.WishlistID == wishlistID {
			wishlistItems = append(wishlistItems, item)
		}
	}

	c.JSON(http.StatusOK, wishlistItems)
}

func updateItem(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")
	itemID := c.Param("item_id")

	var update Item
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Проверяем существование списка и права доступа
	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	if wishlist.UserID != userID && !hasEditAccess(userID, wishlistID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Проверяем существование элемента
	item, exists := items[itemID]
	if !exists || item.WishlistID != wishlistID {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	// Обновляем поля
	item.Name = update.Name
	item.Description = update.Description
	item.Price = update.Price
	item.Link = update.Link
	item.IsPurchased = update.IsPurchased

	items[itemID] = item

	c.JSON(http.StatusOK, item)
}

func deleteItem(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")
	itemID := c.Param("item_id")

	mu.Lock()
	defer mu.Unlock()

	// Проверяем существование списка и права доступа
	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	if wishlist.UserID != userID && !hasEditAccess(userID, wishlistID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Проверяем существование элемента
	item, exists := items[itemID]
	if !exists || item.WishlistID != wishlistID {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	delete(items, itemID)
	c.Status(http.StatusNoContent)
}

func shareWishlist(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	wishlistID := c.Param("id")

	var shareRequest struct {
		SharedUserID string `json:"shared_user_id" binding:"required"`
		CanEdit      bool   `json:"can_edit"`
	}

	if err := c.ShouldBindJSON(&shareRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Проверяем существование списка
	wishlist, exists := wishlists[wishlistID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "wishlist not found"})
		return
	}

	// Проверяем, что пользователь является владельцем
	if wishlist.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can share wishlist"})
		return
	}

	// Проверяем существование пользователя, с которым делимся
	_, exists = users[shareRequest.SharedUserID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to share with not found"})
		return
	}

	// Проверяем, не делимся ли с самим собой
	if shareRequest.SharedUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot share with yourself"})
		return
	}

	// Создаем запись о совместном доступе
	share := SharedWishlist{
		ID:         uuid.New().String(),
		WishlistID: wishlistID,
		UserID:     shareRequest.SharedUserID,
		CanEdit:    shareRequest.CanEdit,
	}

	sharedWishlists[share.ID] = share

	c.JSON(http.StatusCreated, share)
}

func getSharedWishlists(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	mu.RLock()
	defer mu.RUnlock()

	var shared []struct {
		Wishlist Wishlist `json:"wishlist"`
		CanEdit  bool     `json:"can_edit"`
	}

	for _, share := range sharedWishlists {
		if share.UserID == userID {
			if wishlist, exists := wishlists[share.WishlistID]; exists {
				shared = append(shared, struct {
					Wishlist Wishlist `json:"wishlist"`
					CanEdit  bool     `json:"can_edit"`
				}{
					Wishlist: wishlist,
					CanEdit:  share.CanEdit,
				})
			}
		}
	}

	c.JSON(http.StatusOK, shared)
}

// Вспомогательные функции
func hasSharedAccess(userID, wishlistID string) bool {
	for _, share := range sharedWishlists {
		if share.UserID == userID && share.WishlistID == wishlistID {
			return true
		}
	}
	return false
}

func hasEditAccess(userID, wishlistID string) bool {
	for _, share := range sharedWishlists {
		if share.UserID == userID && share.WishlistID == wishlistID && share.CanEdit {
			return true
		}
	}
	return false
}
