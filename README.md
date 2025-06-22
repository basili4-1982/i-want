Набор `curl` запросов для тестирования всех эндпоинтов вашего сервиса списков желаний:

### 1. Регистрация пользователя

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1", "email":"user1@example.com", "password":"password123"}'
```

### 2. Вход пользователя (получение токена)

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1", "password":"password123"}'
```

(Сохраните полученный токен для следующих запросов в переменную `TOKEN`)

### 3. Создание списка желаний

```bash
export TOKEN="your_token_from_login"  # замените на реальный токен
curl -X POST http://localhost:8080/api/wishlists \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"title":"Мой список желаний", "description":"Что я хочу на день рождения"}'
```

(Сохраните ID созданного списка в переменную `WISHLIST_ID`)

### 4. Получение всех списков пользователя

```bash
curl -X GET http://localhost:8080/api/wishlists \
  -H "Authorization: $TOKEN"
```

### 5. Получение конкретного списка

```bash
export WISHLIST_ID="your_wishlist_id"  # замените на реальный ID
curl -X GET http://localhost:8080/api/wishlists/$WISHLIST_ID \
  -H "Authorization: $TOKEN"
```

### 6. Обновление списка

```bash
curl -X PUT http://localhost:8080/api/wishlists/$WISHLIST_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"title":"Обновленный список", "description":"Новое описание"}'
```

### 7. Добавление элемента в список

```bash
curl -X POST http://localhost:8080/api/wishlists/$WISHLIST_ID/items \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"name":"Новый iPhone", "description":"Последняя модель", "price":"999.99", "link":"https://apple.com/iphone"}'
```

(Сохраните ID созданного элемента в переменную `ITEM_ID`)

### 8. Получение всех элементов списка

```bash
curl -X GET http://localhost:8080/api/wishlists/$WISHLIST_ID/items \
  -H "Authorization: $TOKEN"
```

### 9. Обновление элемента

```bash
export ITEM_ID="your_item_id"  # замените на реальный ID
curl -X PUT http://localhost:8080/api/wishlists/$WISHLIST_ID/items/$ITEM_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"name":"iPhone 15 Pro", "price":"1099.99", "isPurchased":false}'
```

### 10. Удаление элемента

```bash
curl -X DELETE http://localhost:8080/api/wishlists/$WISHLIST_ID/items/$ITEM_ID \
  -H "Authorization: $TOKEN"
```

### 11. Регистрация второго пользователя (для теста общего доступа)

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user2", "email":"user2@example.com", "password":"password123"}'
```

### 12. Вход второго пользователя

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user2", "password":"password123"}'
```

(Сохраните токен второго пользователя в переменную `TOKEN2` и его ID в `USER2_ID`)

### 13. Предоставление доступа к списку другому пользователю

```bash
export USER2_ID="id_second_user"  # замените на реальный ID второго пользователя
curl -X POST http://localhost:8080/api/wishlists/$WISHLIST_ID/share \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"shared_user_id":"'$USER2_ID'", "can_edit":true}'
```

### 14. Получение общих списков (для второго пользователя)

```bash
curl -X GET http://localhost:8080/api/shared \
  -H "Authorization: $TOKEN2"
```

### 15. Удаление списка желаний

```bash
curl -X DELETE http://localhost:8080/api/wishlists/$WISHLIST_ID \
  -H "Authorization: $TOKEN"
```

### Примечания:

1. Замените все значения в кавычках (your_token_from_login, your_wishlist_id и т.д.) на реальные значения, полученные из
   ответов сервера.
2. Для удобства тестирования можно сохранять ID в переменные окружения (как показано в примерах).
3. Все запросы требуют заголовка Authorization с токеном, кроме /auth/register и /auth/login.
4. Сервер должен быть запущен на localhost:8080 (или измените URL в запросах, если используете другой адрес).