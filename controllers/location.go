package controllers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "backend-event/models"
    "backend-event/database"
)

// Create Location
func CreateLocation(c *gin.Context) {
    var location models.Location
    if err := c.ShouldBindJSON(&location); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    if err := database.DB.Create(&location).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create location"})
        return
    }

    c.JSON(http.StatusCreated, location)
}

// Get All Locations
func GetAllLocations(c *gin.Context) {
    var locations []models.Location
    if err := database.DB.Find(&locations).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve locations"})
        return
    }

    c.JSON(http.StatusOK, locations)
}

// Get Location by ID
func GetLocationByID(c *gin.Context) {
    id := c.Param("id")
    var location models.Location

    if err := database.DB.First(&location, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
        return
    }

    c.JSON(http.StatusOK, location)
}

// Update Location
func UpdateLocation(c *gin.Context) {
    id := c.Param("id")
    var location models.Location

    if err := database.DB.First(&location, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
        return
    }

    var input struct {
        City string `json:"city"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    location.City = input.City
    if err := database.DB.Save(&location).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location"})
        return
    }

    c.JSON(http.StatusOK, location)
}

// Delete Location
func DeleteLocation(c *gin.Context) {
    id := c.Param("id")
    if err := database.DB.Delete(&models.Location{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete location"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Location deleted successfully"})
}
