package controllers

import (
	"backend-event/database"
	"backend-event/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

func CreateRating(c *gin.Context) {
	var input struct {
		EventID uint `json:"event_id"`
		Rating  int  `json:"rating"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if input.Rating < 1 || input.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating harus antara 1 dan 5"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terotorisasi"})
		return
	}

	loggedInUser, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Data user tidak valid"})
		return
	}

	var existingRating models.Rating
	if err := database.DB.Where("user_id = ? AND event_id = ?", loggedInUser.ID, input.EventID).First(&existingRating).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah memberikan rating untuk event ini"})
		return
	}

	rating := models.Rating{
		UserID:  loggedInUser.ID,
		EventID: input.EventID,
		Rating:  input.Rating,
	}
	if err := database.DB.Create(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memberikan rating"})
		return
	}
	if err := UpdatePopularityScore(input.EventID); err != nil {
		log.Printf("Gagal memperbarui popularity score: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating berhasil diberikan"})
}

func GetEventRatings(c *gin.Context) {
	eventID := c.Param("event_id")

	var ratings []models.Rating
	if err := database.DB.Where("event_id = ?", eventID).Find(&ratings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tidak ada rating untuk event ini"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ratings": ratings})
}

func UpdateRating(c *gin.Context) {
	var input struct {
		EventID uint `json:"event_id"`
		Rating  int  `json:"rating"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if input.Rating < 1 || input.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating harus antara 1 dan 5"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terotorisasi"})
		return
	}

	loggedInUser, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Data user tidak valid"})
		return
	}

	var existingRating models.Rating
	if err := database.DB.Where("user_id = ? AND event_id = ?", loggedInUser.ID, input.EventID).First(&existingRating).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rating tidak ditemukan"})
		return
	}

	existingRating.Rating = input.Rating
	if err := database.DB.Save(&existingRating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating berhasil diperbarui"})
}

func DeleteRating(c *gin.Context) {
	var input struct {
		EventID uint `json:"event_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terotorisasi"})
		return
	}

	loggedInUser, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Data user tidak valid"})
		return
	}

	var existingRating models.Rating
	if err := database.DB.Where("user_id = ? AND event_id = ?", loggedInUser.ID, input.EventID).First(&existingRating).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rating tidak ditemukan"})
		return
	}

	if err := database.DB.Delete(&existingRating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating berhasil dihapus"})
}
