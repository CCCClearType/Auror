package routes

import (
	"auror_vapor_backend/controllers"
	"auror_vapor_backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 40 << 20 // 40 MB
	r.Static("/media/images", "./assets/images")

	api := r.Group("/api")
	{
		// ==========================================
		// Phase 2: Auth Routes (Public)
		// ==========================================
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/logout", middleware.RequireAuth(), controllers.Logout)
		}

		// ==========================================
		// Phase 1: Note Routes (Public)
		// ==========================================
		api.GET("/notes", controllers.GetNotes)
		api.GET("/notes/:id", controllers.GetNoteByID)
		api.GET("/notes/:id/reviews", controllers.GetReviews)
		
		api.GET("/ilearn-status", controllers.CheckIlearnStatus)
		api.POST("/ilearn-reports", controllers.SubmitIlearnReport)
		api.GET("/ilearn-history", controllers.GetIlearnHistory)

		// ==========================================
		// Phase 3: Protected Routes (Require JWT Token)
		// ==========================================
		protected := api.Group("/protected").Use(middleware.RequireAuth())
		{
			// 1. Shopping Cart
			protected.GET("/cart", controllers.GetCart)
			protected.POST("/cart", controllers.AddToCart)
			protected.DELETE("/cart/:note_id", controllers.RemoveFromCart)

			// 2. Checkout & Transactions
			protected.POST("/checkout", controllers.Checkout)
			protected.GET("/transactions", controllers.GetTransactions)
			protected.GET("/refunds", controllers.GetMyRefunds)

			// 3. Library & Wishlist
			protected.GET("/library", controllers.GetLibrary)
			protected.GET("/wishlist", controllers.GetWishlist)
			protected.POST("/wishlist", controllers.AddToWishlist)
			protected.DELETE("/wishlist/:note_id", controllers.RemoveFromWishlist)
			protected.GET("/library/:note_id/play", controllers.PlayNote)
			protected.GET("/library/:note_id/download", controllers.DownloadNote)
		}

		// ==========================================
		// Phase 4: Seller Routes (Require JWT + SELLER Role)
		// ==========================================
		seller := api.Group("/seller").Use(middleware.RequireAuth(), middleware.RequireRole("SELLER"))
		{
			seller.GET("/notes", controllers.GetSellerNotes)
			seller.POST("/notes", controllers.UploadNote)
			seller.PUT("/notes/:id", controllers.UpdateNote)
			seller.PUT("/notes/:id/publish", controllers.PublishNote)
			seller.DELETE("/notes/:id", controllers.DeleteNote)
			seller.POST("/notes/:id/media", controllers.UploadMedia)
			seller.DELETE("/notes/:id/media/:media_id", controllers.DeleteMedia)
			seller.GET("/notes/:id/stats", controllers.GetNoteStats)
			seller.POST("/tags", controllers.CreateTag)
			seller.POST("/notes/:id/tags", controllers.AddTagToNote)
			seller.DELETE("/notes/:id/tags/:tag_id", controllers.RemoveTagFromNote)
		}

		api.GET("/tags", controllers.GetTags)

		// ==========================================
		// Phase 5: Complete Remaining API Groups
		// ==========================================

		// Protected User Routes
		users := api.Group("/users").Use(middleware.RequireAuth())
		{
			users.PUT("/profile", controllers.UpdateProfile)
		}

		// ADMIN Zone
		admin := api.Group("/admin").Use(middleware.RequireAuth(), middleware.RequireRole("ADMIN"))
		{
			admin.GET("/users", controllers.GetUsers)
			admin.PUT("/users/:id/suspend", controllers.SuspendUser)
			admin.DELETE("/users/:id", controllers.DeleteUser)
			admin.PUT("/users/:id/role", controllers.ChangeUserRole)
			admin.DELETE("/notes/:id", controllers.AdminDeleteNote)
		}

		// CSR Zone
		csr := api.Group("/csr").Use(middleware.RequireAuth(), middleware.RequireRole("CSR"))
		{
			csr.GET("/refunds", controllers.GetRefundRequests)
			csr.PUT("/refunds/:id", controllers.ProcessRefund)
		}

		// SOCIAL Zone
		social := api.Group("/social").Use(middleware.RequireAuth())
		{
			social.POST("/notes/:id/reviews", controllers.PostReview)
			social.POST("/reviews/:review_id/replies", controllers.ReplyToReview)
			social.DELETE("/reviews/replies/:reply_id", controllers.DeleteReviewReply)
			social.POST("/refunds", controllers.ApplyRefund)

			// Friends
			social.GET("/friends", controllers.GetFriends)
			social.GET("/friends/requests", controllers.GetFriendRequests)
			social.POST("/friends/request", controllers.SendFriendRequest)
			social.PUT("/friends/request/:id/accept", controllers.AcceptFriendRequest)
			social.PUT("/friends/request/:id/decline", controllers.DeclineFriendRequest)
			social.DELETE("/friends/request/:id", controllers.RevokeFriendRequest)

			// Blacklist
			social.GET("/blacklist", controllers.GetBlacklist)
			social.POST("/blacklist", controllers.AddBlacklist)
			social.DELETE("/blacklist/:user_id", controllers.RemoveBlacklist)

			// Messages
			social.POST("/messages", controllers.SendMessage)
			social.GET("/messages/:user_id", controllers.GetMessages)
		}
	}

	return r
}
