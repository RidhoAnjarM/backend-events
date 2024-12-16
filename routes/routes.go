package routes

import (
	"backend-event/controllers"
	"backend-event/middlewares"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	router := r.Group("/api")
	{
		router.POST("/login", controllers.Login)
		router.POST("/register", controllers.Register)
		router.GET("/profile", middlewares.AuthMiddleware(), controllers.GetProfile)

		// user
		router.GET("/user", controllers.GetAllUsers)
		router.GET("/user/:id", controllers.GetUserById)
		router.PUT("/user/:id", controllers.UpdateUser)
		router.DELETE("user/:id", controllers.DeleteUser)

		// event
		router.POST("/event", controllers.CreateEvent)
		router.GET("/event", controllers.GetAllEvents)
		router.GET("/event/:id", controllers.GetEventByID)
		router.PUT("/event/:id", controllers.UpdateEvent)
		router.DELETE("/event/:id", controllers.DeleteEvent)

		// daftar event
		router.POST("/events/:event_id/register", middlewares.AuthMiddleware(), controllers.RegisterEvent)
		router.GET("/events/registered", middlewares.AuthMiddleware(), controllers.GetRegisteredEvents)
		router.GET("/events/:event_id/registered", controllers.GetEventRegistrants)

		router.GET("/events/:event_id/check-registration", middlewares.AuthMiddleware(), controllers.CheckRegistration)

		//kategori
		router.POST("/categories", controllers.CreateCategory)
		router.GET("/categories", controllers.GetCategories)
		router.GET("/categories/:id", controllers.GetCategoryByID)
		router.PUT("/categories/:id", controllers.UpdateCategory)
		router.DELETE("/categories/:id", controllers.DeleteCategory)

		//lokasi
		router.POST("/location", controllers.CreateLocation)
		router.GET("/location", controllers.GetAllLocations)
		router.GET("/location/:id", controllers.GetLocationByID)
		router.PUT("/location/:id", controllers.UpdateLocation)
		router.DELETE("/location/:id", controllers.DeleteLocation)

		//rating
		router.POST("/rating", middlewares.AuthMiddleware(), controllers.CreateRating)
		router.GET("/events/:event_id/ratings", middlewares.AuthMiddleware(), controllers.GetEventRatings)
		router.PUT("/rating/:id", middlewares.AuthMiddleware(), controllers.UpdateRating)
		router.DELETE("/rating/:id", middlewares.AuthMiddleware(), controllers.DeleteRating)

		router.GET("/events/populars", controllers.GetPopularEvents)
		router.GET("/events/unregister",  middlewares.AuthMiddleware(), controllers.GetUnregisteredEvents)
	}
}
