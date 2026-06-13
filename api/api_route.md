# API 規格與後端伺服器 (Go Server) 路由稽核報告

經過詳細的代碼追蹤與交叉比對，我將 API 規格書中的 53 支 API，與實際運行在 Go 後端伺服器的 `backend/routes/routes.go` 進行了一對一的比對，確認實作狀態。

## 📝 實作細節盤點

### 1. 公開路由 (Public Routes) - 免 JWT
| API 規格 | `routes.go` 註冊狀態 | 驗證 |
| :--- | :--- | :---: |
| `[POST] /api/auth/register` | `auth.POST("/register")` | ✅ |
| `[POST] /api/auth/login` | `auth.POST("/login")` | ✅ |
| `[GET] /api/notes` | `api.GET("/notes")` | ✅ |
| `[GET] /api/notes/{id}` | `api.GET("/notes/:id")` | ✅ |
| `[GET] /api/notes/{id}/reviews` | `api.GET("/notes/:id/reviews")` | ✅ |
| `[GET] /api/tags` | `api.GET("/tags")` | ✅ |

### 2. 一般登入保護 (Protected Routes) - 僅需 JWT
*這包含購物車、買家筆記庫、個人檔案，以及社群互動等。*
| API 規格 | `routes.go` 註冊狀態 | 驗證 |
| :--- | :--- | :---: |
| `[POST] /api/auth/logout` | `auth.POST("/logout")` | ✅ |
| `[PUT] /api/users/profile` | `users.PUT("/profile")` | ✅ |
| `[GET] /api/protected/cart` | `protected.GET("/cart")` | ✅ |
| `[POST] /api/protected/cart` | `protected.POST("/cart")` | ✅ |
| `[DELETE] /api/protected/cart/{id}` | `protected.DELETE("/cart/:note_id")` | ✅ |
| `[POST] /api/protected/checkout` | `protected.POST("/checkout")` | ✅ |
| `[GET] /api/protected/transactions` | `protected.GET("/transactions")` | ✅ |
| `[GET] /api/protected/refunds` | `protected.GET("/refunds")` | ✅ |
| `[GET] /api/protected/library` | `protected.GET("/library")` | ✅ |
| `[GET] /api/protected/wishlist` | `protected.GET("/wishlist")` | ✅ |
| `[POST] /api/protected/wishlist` | `protected.POST("/wishlist")` | ✅ |
| `[DELETE] /api/protected/wishlist/{id}` | `protected.DELETE("/wishlist/:note_id")` | ✅ |
| `[GET] /api/protected/library/.../play` | `protected.GET("/library/:note_id/play")` | ✅ |
| `[GET] /api/protected/library/.../download` | `protected.GET("/library/:note_id/download")` | ✅ |
| `[POST] /api/social/.../reviews` | `social.POST("/notes/:id/reviews")` | ✅ |
| `[POST] /api/social/.../replies` | `social.POST("/reviews/:review_id/replies")` | ✅ |
| `[DELETE] /api/social/.../replies/{id}`| `social.DELETE("/reviews/replies/:reply_id")` | ✅ |
| `[POST] /api/social/refunds` | `social.POST("/refunds")` | ✅ |
| `[GET] /api/social/friends` | `social.GET("/friends")` | ✅ |
| `[GET] /api/social/friends/requests` | `social.GET("/friends/requests")` | ✅ |
| `[POST] /api/social/friends/request` | `social.POST("/friends/request")` | ✅ |
| `[PUT] /api/social/friends/.../accept` | `social.PUT("/friends/request/:id/accept")` | ✅ |
| `[PUT] /api/social/friends/.../decline`| `social.PUT("/friends/request/:id/decline")`| ✅ |
| `[DELETE] /api/social/friends/request/{id}` | `social.DELETE("/friends/request/:id")` | ✅ |
| `[GET] /api/social/blacklist` | `social.GET("/blacklist")` | ✅ |
| `[POST] /api/social/blacklist` | `social.POST("/blacklist")` | ✅ |
| `[DELETE] /api/social/blacklist/{id}` | `social.DELETE("/blacklist/:user_id")` | ✅ |
| `[POST] /api/social/messages` | `social.POST("/messages")` | ✅ |
| `[GET] /api/social/messages/{user_id}` | `social.GET("/messages/:user_id")` | ✅ |

### 3. 賣家路由 (Seller Routes) - 需 SELLER 權限
| API 規格 | `routes.go` 註冊狀態 | 驗證 |
| :--- | :--- | :---: |
| `[GET] /api/seller/notes` | `seller.GET("/notes")` | ✅ |
| `[POST] /api/seller/notes` | `seller.POST("/notes")` | ✅ |
| `[PUT] /api/seller/notes/{id}/publish` | `seller.PUT("/notes/:id/publish")` | ✅ |
| `[PUT] /api/seller/notes/{id}` | `seller.PUT("/notes/:id")` | ✅ |
| `[DELETE] /api/seller/notes/{id}` | `seller.DELETE("/notes/:id")` | ✅ |
| `[POST] /api/seller/notes/{id}/media` | `seller.POST("/notes/:id/media")` | ✅ |
| `[DELETE] /api/seller/.../media/{id}` | `seller.DELETE("/notes/:id/media/:media_id")` | ✅ |
| `[GET] /api/seller/notes/{id}/stats` | `seller.GET("/notes/:id/stats")` | ✅ |
| `[POST] /api/seller/tags` | `seller.POST("/tags")` | ✅ |
| `[POST] /api/seller/notes/{id}/tags` | `seller.POST("/notes/:id/tags")` | ✅ |
| `[DELETE] /api/seller/.../tags/{id}` | `seller.DELETE("/notes/:id/tags/:tag_id")` | ✅ |

### 4. 管理員與客服路由 (Admin & CSR Routes)
| API 規格 | `routes.go` 註冊狀態 | 驗證 |
| :--- | :--- | :---: |
| `[GET] /api/admin/users` | `admin.GET("/users")` | ✅ |
| `[PUT] /api/admin/users/{id}/suspend` | `admin.PUT("/users/:id/suspend")` | ✅ |
| `[DELETE] /api/admin/users/{id}` | `admin.DELETE("/users/:id")` | ✅ |
| `[PUT] /api/admin/users/{id}/role` | `admin.PUT("/users/:id/role")` | ✅ |
| `[DELETE] /api/admin/notes/{id}` | `admin.DELETE("/notes/:id")` | ✅ |
| `[GET] /api/csr/refunds` | `csr.GET("/refunds")` | ✅ |
| `[PUT] /api/csr/refunds/{id}` | `csr.PUT("/refunds/:id")` | ✅ |

---
> [!NOTE]
> **補充說明**
> 1. API 文件中的路徑參數採用標準的 `{id}` 寫法，在 Go/Gin 框架中實作為 `:id`（如 `:note_id` 或 `:user_id`），此為框架約定俗成的語法對應，邏輯上等同於文件描述。
> 2. `auth`, `admin`, `csr`, `seller`, `protected`, `social` 各大 Router Group 皆已正確綁定對應的 `RequireRole()` 或 `RequireAuth()` 中介軟體防護。
