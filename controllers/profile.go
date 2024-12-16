package controllers

import ( 
	"backend-event/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	loggedInUser := user.(models.User)
 
	c.JSON(http.StatusOK, gin.H{
		"id": loggedInUser.ID,
		"username": loggedInUser.Username,
	})
}
