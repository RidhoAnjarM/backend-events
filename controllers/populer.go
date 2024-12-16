package controllers

import (
	"backend-event/database"
	"backend-event/models"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
)

func UpdatePopularityScore(eventID uint) error {
	var totalRegistrations int64
	var averageRating float64

	if err := database.DB.Model(&models.Registration{}).Where("event_id = ?", eventID).Count(&totalRegistrations).Error; err != nil {
		return err
	}

	if err := database.DB.Model(&models.Rating{}).Where("event_id = ?", eventID).Select("COALESCE(AVG(rating), 0)").Scan(&averageRating).Error; err != nil {
		return err
	}

	popularityScore := float64(totalRegistrations)*0.7 + averageRating*0.3

	if err := database.DB.Model(&models.Event{}).Where("id = ?", eventID).Update("popularity_score", popularityScore).Error; err != nil {
		return err
	}

	return nil
}

func GetPopularEvents(c *gin.Context) {
	var events []models.Event

	if err := database.DB.Order("popularity_score DESC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar event populer"})
		return
	}

	var response []struct {
		ID                uint    `json:"id"`
		Name              string  `json:"name"`
		Description       string  `json:"description"`
		DateStart         string  `json:"date_start"`
		DateEnd           string  `json:"date_end"`
		Time              string  `json:"time"`
		Location          string  `json:"location"`
		Address           string  `json:"address"`
		Capacity          int     `json:"capacity"`
		RemainingCapacity int     `json:"remaining_capacity"`
		Photo             string  `json:"photo"`
		Price             string  `json:"price"`
		Category          string  `json:"category"`
		AverageRating     float64 `json:"average_rating"`
		UniqueRaters      int64   `json:"unique_raters"`
		PopularityScore   float64 `json:"popularity_score"`
	}

	for _, event := range events {
		var location models.Location
		if err := database.DB.First(&location, event.LocationID).Error; err != nil {
			location.City = ""
		}

		var category models.Category
		if err := database.DB.First(&category, event.CategoryID).Error; err != nil {
			category.Name = ""
		}

		var averageRating float64
		if err := database.DB.Model(&models.Rating{}).
			Where("event_id = ?", event.ID).
			Select("COALESCE(AVG(rating), 0)").Scan(&averageRating).Error; err != nil {
			averageRating = 0.0
		}

		averageRating = math.Round(averageRating*100) / 100

		var uniqueRaters int64
		if err := database.DB.Model(&models.Rating{}).
			Where("event_id = ?", event.ID).
			Select("COUNT(DISTINCT user_id)").Scan(&uniqueRaters).Error; err != nil {
			uniqueRaters = 0
		}

		response = append(response, struct {
			ID                uint    `json:"id"`
			Name              string  `json:"name"`
			Description       string  `json:"description"`
			DateStart         string  `json:"date_start"`
			DateEnd           string  `json:"date_end"`
			Time              string  `json:"time"`
			Location          string  `json:"location"`
			Address           string  `json:"address"`
			Capacity          int     `json:"capacity"`
			RemainingCapacity int     `json:"remaining_capacity"`
			Photo             string  `json:"photo"`
			Price             string  `json:"price"`
			Category          string  `json:"category"`
			AverageRating     float64 `json:"average_rating"`
			UniqueRaters      int64   `json:"unique_raters"`
			PopularityScore   float64 `json:"popularity_score"`
		}{
			ID:                event.ID,
			Name:              event.Name,
			Description:       event.Description,
			DateStart:         event.DateStart,
			DateEnd:           event.DateEnd,
			Time:              event.Time,
			Location:          location.City,
			Address:           event.Address,
			Capacity:          event.Capacity,
			RemainingCapacity: event.RemainingCapacity,
			Photo:             event.Photo,
			Price:             event.Price,
			Category:          category.Name,
			AverageRating:     averageRating,
			UniqueRaters:      uniqueRaters,
			PopularityScore:   event.PopularityScore,
		})
	}

	c.JSON(http.StatusOK, response)
}
