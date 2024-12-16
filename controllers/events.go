package controllers

import (
	"backend-event/database"
	"backend-event/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

func CreateEvent(c *gin.Context) {
	var event models.Event

	file, err := c.FormFile("photo")
	if err != nil {
		event.Photo = ""
	} else {
		uploadPath := fmt.Sprintf("./uploads/%s", file.Filename)

		if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
			if err := os.Mkdir("./uploads", os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
				return
			}
		}

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo"})
			return
		}

		event.Photo = fmt.Sprintf("/uploads/%s", file.Filename)
	}

	event.Name = c.PostForm("name")
	event.Description = c.PostForm("description")
	event.DateStart = c.PostForm("datestart")
	event.DateEnd = c.PostForm("dateend")
	event.Time = c.PostForm("time")
	event.Benefits = c.PostForm("benefits")
	event.Mode = c.PostForm("mode")
	event.Link = c.PostForm("link")
	event.Price = c.PostForm("price")
	if event.Price == "" {
		event.Price = "Free"
	}

	capacity, err := strconv.Atoi(c.PostForm("capacity"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid capacity"})
		return
	}
	event.Capacity = capacity
	event.RemainingCapacity = capacity

	locationID, err := strconv.Atoi(c.PostForm("location_id"))
	if err != nil || locationID == 0 {
		event.Mode = "online"
		event.LocationID = 0
		event.Location = "Online"
		event.Address = ""
	} else {
		event.LocationID = uint(locationID)
		var location models.Location
		if err := database.DB.First(&location, event.LocationID).Error; err == nil {
			event.Location = location.City
			event.Address = c.PostForm("address")
			if event.Address == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Address is required when location is specified"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
			return
		}
	}

	categoryID, err := strconv.Atoi(c.PostForm("category_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	event.CategoryID = uint(categoryID)

	if event.Mode == "online" && event.Link == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link is required for online events"})
		return
	}

	dateStart, err := time.Parse("2006-01-02", event.DateStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
		return
	}

	var dateEnd time.Time
	if event.DateEnd != "" {
		dateEnd, err = time.Parse("2006-01-02", event.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		if dateEnd.Before(dateStart) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End date cannot be before start date"})
			return
		}
	}

	currentDate := time.Now()
	if currentDate.Before(dateStart) {
		event.Status = "upcoming"
	} else if currentDate.After(dateStart) && (event.DateEnd == "" || currentDate.Before(dateEnd)) {
		event.Status = "ongoing"
	} else if event.DateEnd != "" && currentDate.After(dateEnd) {
		event.Status = "ended"
	}

	if err := database.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event", "details": err.Error()})
		return
	}

	var sessions []models.Session
	for i := 0; ; i++ {
		date := c.PostForm(fmt.Sprintf("sessions[%d][date]", i))
		sessionTime := c.PostForm(fmt.Sprintf("sessions[%d][time]", i))
		speaker := c.PostForm(fmt.Sprintf("sessions[%d][speaker]", i))
		location := c.PostForm(fmt.Sprintf("sessions[%d][location]", i))

		if date == "" {
			break
		}

		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Format tanggal tidak valid untuk sesi %d", i)})
			return
		}

		sessions = append(sessions, models.Session{
			EventID:  event.ID,
			Date:     date,
			Time:     sessionTime,
			Speaker:  speaker,
			Location: location,
		})
	}

	if len(sessions) > 0 {
		if err := database.DB.Create(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sessions", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Event and sessions created successfully",
		"event":    event,
		"sessions": sessions,
	})
}

// get semua event
func GetAllEvents(c *gin.Context) {
	var events []models.Event
	if err := database.DB.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
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
		Status            string  `json:"status"`
		Mode              string  `json:"mode"`
		AverageRating     float64 `json:"average_rating"`
		UniqueRaters      int64   `json:"unique_raters"`
	}

	for _, event := range events {
		var location models.Location
		if event.LocationID != 0 {
			if err := database.DB.First(&location, event.LocationID).Error; err != nil {
				location.City = ""
			}
		} else {
			location.City = "Online"
		}

		var category models.Category
		if err := database.DB.First(&category, event.CategoryID).Error; err != nil {
			category.Name = ""
		}

		var eventStatus string
		dateStart, err := time.Parse("2006-01-02", event.DateStart)
		if err != nil {
			eventStatus = "unknown"
		} else {
			currentDate := time.Now()

			if currentDate.Before(dateStart) {
				eventStatus = "upcoming"
			} else {
				if event.DateEnd != "" {
					dateEnd, err := time.Parse("2006-01-02", event.DateEnd)
					if err != nil {
						eventStatus = "unknown"
					} else if currentDate.Before(dateEnd) {
						eventStatus = "ongoing"
					} else {
						eventStatus = "ended"
					}
				} else {
					currentDateStr := currentDate.Format("2006-01-02")
					currentDateOnly, _ := time.Parse("2006-01-02", currentDateStr)
					dateStartOnly, _ := time.Parse("2006-01-02", event.DateStart)
					
					if currentDateOnly.After(dateStartOnly) {
						eventStatus = "ended"
					} else {
						eventStatus = "ongoing"
					}
				}
			}
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
			Status            string  `json:"status"`
			Mode              string  `json:"mode"`
			AverageRating     float64 `json:"average_rating"`
			UniqueRaters      int64   `json:"unique_raters"`
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
			Status:            eventStatus,
			Mode:              event.Mode,
			AverageRating:     averageRating,
			UniqueRaters:      uniqueRaters,
		})
	}

	c.JSON(http.StatusOK, response)
}

// get event per id
func GetEventByID(c *gin.Context) {
	id := c.Param("id")

	var event models.Event
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var location models.Location
	if event.LocationID != 0 {
		if err := database.DB.First(&location, event.LocationID).Error; err != nil {
			location.City = ""
		}
	} else {
		location.City = "Online"
	}

	var category models.Category
	if err := database.DB.First(&category, event.CategoryID).Error; err != nil {
		category.Name = ""
	}

	var eventStatus string
	dateStart, err := time.Parse("2006-01-02", event.DateStart)
	if err != nil {
		eventStatus = "unknown"
	} else {
		currentDate := time.Now()

		if currentDate.Before(dateStart) {
			eventStatus = "upcoming"
		} else {
			if event.DateEnd != "" {
				dateEnd, err := time.Parse("2006-01-02", event.DateEnd)
				if err != nil {
					eventStatus = "unknown"
				} else if currentDate.Before(dateEnd) {
					eventStatus = "ongoing"
				} else {
					eventStatus = "ended"
				}
			} else {
				currentDateStr := currentDate.Format("2006-01-02")
				currentDateOnly, _ := time.Parse("2006-01-02", currentDateStr)
				dateStartOnly, _ := time.Parse("2006-01-02", event.DateStart)
				
				if currentDateOnly.After(dateStartOnly) {
					eventStatus = "ended"
				} else {
					eventStatus = "ongoing"
				}
			}
		}
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

	var eventLink string
	if event.Mode == "online" {
		eventLink = event.Link
	}

	var sessions []models.Session
	if err := database.DB.Where("event_id = ?", event.ID).Find(&sessions).Error; err != nil {
		sessions = []models.Session{}
	}

	response := struct {
		ID                uint             `json:"id"`
		Name              string           `json:"name"`
		Description       string           `json:"description"`
		DateStart         string           `json:"date_start"`
		DateEnd           string           `json:"date_end"`
		Time              string           `json:"time"`
		Location          string           `json:"location"`
		LocationID        uint             `json:"location_id"`
		Address           string           `json:"address"`
		Capacity          int              `json:"capacity"`
		RemainingCapacity int              `json:"remaining_capacity"`
		Photo             string           `json:"photo"`
		Price             string           `json:"price"`
		Category          string           `json:"category"`
		CategoryID        uint             `json:"category_id"`
		Benefits          string           `json:"benefits"`
		Mode              string           `json:"mode"`
		Link              string           `json:"link,omitempty"`
		Status            string           `json:"status"`
		AverageRating     float64          `json:"average_rating"`
		UniqueRaters      int64            `json:"unique_raters"`
		Sessions          []models.Session `json:"sessions"`
	}{
		ID:                event.ID,
		Name:              event.Name,
		Description:       event.Description,
		DateStart:         event.DateStart,
		DateEnd:           event.DateEnd,
		Time:              event.Time,
		Location:          location.City,
		LocationID:        event.LocationID,
		Address:           event.Address,
		Capacity:          event.Capacity,
		RemainingCapacity: event.RemainingCapacity,
		Photo:             event.Photo,
		Price:             event.Price,
		Category:          category.Name,
		CategoryID:        event.CategoryID,
		Benefits:          event.Benefits,
		Mode:              event.Mode,
		Link:              eventLink,
		Status:            eventStatus,
		AverageRating:     averageRating,
		UniqueRaters:      uniqueRaters,
		Sessions:          sessions,
	}

	c.JSON(http.StatusOK, response)
}

// update event
func UpdateEvent(c *gin.Context) {
	var event models.Event

	id := c.Param("id")
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	file, err := c.FormFile("photo")
	if err == nil {
		uploadPath := fmt.Sprintf("./uploads/%s", file.Filename)

		if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
			if err := os.Mkdir("./uploads", os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
				return
			}
		}

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo"})
			return
		}

		event.Photo = fmt.Sprintf("/uploads/%s", file.Filename)
	}

	event.Name = c.PostForm("name")
	event.Description = c.PostForm("description")
	event.DateStart = c.PostForm("datestart")
	event.DateEnd = c.PostForm("dateend")
	event.Time = c.PostForm("time")
	event.Benefits = c.PostForm("benefits")
	event.Mode = c.PostForm("mode")
	event.Link = c.PostForm("link")
	event.Price = c.PostForm("price")
	if event.Price == "" {
		event.Price = "Free"
	}

	capacity, err := strconv.Atoi(c.PostForm("capacity"))
	if err == nil {
		 oldCapacity := event.Capacity
		
		event.Capacity = capacity
		
		if capacity < oldCapacity {
			event.RemainingCapacity = capacity
		}
	}

	remainingCapacity, err := strconv.Atoi(c.PostForm("remaining_capacity"))
	if err == nil {
		if remainingCapacity > event.Capacity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Remaining capacity tidak boleh melebihi capacity"})
			return
		}
		
		if remainingCapacity < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Remaining capacity tidak boleh kurang dari 0"})
			return
		}
		
		event.RemainingCapacity = remainingCapacity
	}

	if err := database.DB.Model(&event).Updates(map[string]interface{}{
		"capacity":           event.Capacity,
		"remaining_capacity": event.RemainingCapacity,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update capacity"})
		return
	}

	locationID, err := strconv.Atoi(c.PostForm("location_id"))
	if err == nil && locationID != 0 {
		event.LocationID = uint(locationID)
		var location models.Location
		if err := database.DB.First(&location, event.LocationID).Error; err == nil {
			event.Location = location.City
			event.Address = c.PostForm("address")
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
			return
		}
	} else {
		event.Mode = "online"
		event.LocationID = 0
		event.Location = "Online"
		event.Address = ""
	}

	categoryID, err := strconv.Atoi(c.PostForm("category_id"))
	if err == nil {
		event.CategoryID = uint(categoryID)
	}

	if event.Mode == "online" && event.Link == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link is required for online events"})
		return
	}

	dateStart, err := time.Parse("2006-01-02", event.DateStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
		return
	}

	if event.DateEnd != "" {
		dateEnd, err := time.Parse("2006-01-02", event.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		if dateEnd.Before(dateStart) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End date cannot be before start date"})
			return
		}
	}

	currentDate := time.Now()
	if event.DateEnd == "" {
		if currentDate.Before(dateStart) {
			event.Status = "upcoming"
		} else {
			event.Status = "ongoing"
		}
	} else {
		dateEnd, err := time.Parse("2006-01-02", event.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		if dateEnd.Before(dateStart) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End date cannot be before start date"})
			return
		}
		if currentDate.Before(dateStart) {
			event.Status = "upcoming"
		} else if currentDate.After(dateStart) && currentDate.Before(dateEnd) {
			event.Status = "ongoing"
		} else {
			event.Status = "ended"
		}
	}

	if err := database.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event", "details": err.Error()})
		return
	}

	var sessions []models.Session
	database.DB.Where("event_id = ?", event.ID).Delete(&sessions)

	for i := 0; ; i++ {
		date := c.PostForm(fmt.Sprintf("sessions[%d][date]", i))
		sessionTime := c.PostForm(fmt.Sprintf("sessions[%d][time]", i))
		speaker := c.PostForm(fmt.Sprintf("sessions[%d][speaker]", i))
		location := c.PostForm(fmt.Sprintf("sessions[%d][location]", i))

		if date == "" {
			break
		}

		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Format tanggal tidak valid untuk sesi %d", i)})
			return
		}

		sessions = append(sessions, models.Session{
			EventID:  event.ID,
			Date:     date,
			Time:     sessionTime,
			Speaker:  speaker,
			Location: location,
		})
	}

	if len(sessions) > 0 {
		if err := database.DB.Create(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sessions", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Event and sessions updated successfully",
		"event":    event,
		"sessions": sessions,
	})
}

// delete event
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")

	var event models.Event
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if err := database.DB.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// ini buat liat event dan siapa aja yang daftar
func GetEventRegistrants(c *gin.Context) {
	eventID := c.Param("event_id")

	id, err := strconv.Atoi(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var userEvents []models.Registration
	if err := database.DB.Where("event_id = ?", id).Find(&userEvents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registrants"})
		return
	}

	var registrants []gin.H
	for _, ue := range userEvents {
		registrants = append(registrants, gin.H{
			"username": ue.Username,
			"name":     ue.Name,
			"email":    ue.Email,
			"phone":    ue.PhoneNumber,
			"job":      ue.Job,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":    event.ID,
		"event_name":  event.Name,
		"registrants": registrants,
	})
}

// ini buat liat user daftar dievent mana aja
func GetRegisteredEvents(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    loggedInUser := user.(models.User)

    var registeredEvents []models.Registration
    if err := database.DB.Preload("Event").Where("user_id = ?", loggedInUser.ID).Find(&registeredEvents).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registered events"})
        return
    }

    currentDate := time.Now()
    var events []gin.H
    for _, ue := range registeredEvents {
        var averageRating float64
        if err := database.DB.Model(&models.Rating{}).
            Where("event_id = ?", ue.Event.ID).
            Select("COALESCE(AVG(rating), 0)").Scan(&averageRating).Error; err != nil {
            averageRating = 0.0
        }
        averageRating = math.Round(averageRating*100) / 100

        var uniqueRaters int64
        if err := database.DB.Model(&models.Rating{}).
            Where("event_id = ?", ue.Event.ID).
            Select("COUNT(DISTINCT user_id)").Scan(&uniqueRaters).Error; err != nil {
            uniqueRaters = 0
        }

        dateStart, err := time.Parse("2006-01-02", ue.Event.DateStart)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
            return
        }

        var status string
        if ue.Event.DateEnd == "" {
            if currentDate.Before(dateStart) {
                status = "akan datang"
            } else {
                status = "sedang berjalan"
            }
        } else {
            dateEnd, err := time.Parse("2006-01-02", ue.Event.DateEnd)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
                return
            }

            if dateEnd.Before(dateStart) {
                c.JSON(http.StatusBadRequest, gin.H{"error": "End date cannot be before start date"})
                return
            }

            if currentDate.Before(dateStart) {
                status = "akan datang"
            } else if currentDate.After(dateStart) && currentDate.Before(dateEnd) {
                status = "sedang berjalan"
            } else {
                status = "selesai"
            }
        }

        events = append(events, gin.H{
            "id":           ue.Event.ID,
            "name":         ue.Event.Name,
            "description":  ue.Event.Description,
            "date":         ue.Event.DateStart,
            "time":         ue.Event.Time,
            "location":     ue.Event.Location,
            "address":      ue.Event.Address,
            "capacity":     ue.Event.Capacity,
            "photo":        ue.Event.Photo,
            "price":        ue.Event.Price,
            "name_reg":     ue.Name,
            "email":        ue.Email,
            "phone":        ue.PhoneNumber,
            "job":          ue.Job,
            "status":       status,   
            "rating":       averageRating,
            "uniqueraters": uniqueRaters,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "username":          loggedInUser.Username,
        "registered_events": events,
    })
}

//ini buat cek user udah daftar dievent
func CheckRegistration(c *gin.Context) {
	eventID := c.Param("event_id")

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

	var registration models.Registration
	if err := database.DB.Where("user_id = ? AND event_id = ?", loggedInUser.ID, eventID).First(&registration).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"isRegistered": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isRegistered": false})
}

//buat cek user belum daftar event
func GetUnregisteredEvents(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	loggedInUser := user.(models.User)

	var registeredEventIDs []uint
	if err := database.DB.Model(&models.Registration{}).
		Where("user_id = ?", loggedInUser.ID).
		Pluck("event_id", &registeredEventIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registered events"})
		return
	}

	var eventsToDisplay []models.Event

	if len(registeredEventIDs) == 0 {
		if err := database.DB.Where("status IN ?", []string{"upcoming", "ongoing"}).
			Limit(6).Find(&eventsToDisplay).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
			return
		}
	} else {
		if err := database.DB.Where("id NOT IN ? AND status IN ?", registeredEventIDs, []string{"upcoming", "ongoing"}).
			Limit(6).Find(&eventsToDisplay).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unregistered events"})
			return
		}
	}

	var events []gin.H
	currentDate := time.Now()

	for _, event := range eventsToDisplay {
		if event.DateEnd == "" {
			dateStart, err := time.Parse("2006-01-02", event.DateStart)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
				return
			}
			if currentDate.Before(dateStart) {
				event.Status = "upcoming"
			} else {
				event.Status = "ongoing"
			}
		} else {
			dateStart, err := time.Parse("2006-01-02", event.DateStart)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
				return
			}

			dateEnd, err := time.Parse("2006-01-02", event.DateEnd)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
				return
			}

			if currentDate.Before(dateStart) {
				event.Status = "upcoming"
			} else if currentDate.After(dateStart) && currentDate.Before(dateEnd) {
				event.Status = "ongoing"
			} else {
				event.Status = "ended"
			}
		}

		if event.Status == "upcoming" || event.Status == "ongoing" {
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

			events = append(events, gin.H{
				"id":             event.ID,
				"name":           event.Name,
				"description":    event.Description,
				"date":           event.DateStart,
				"time":           event.Time,
				"location":       event.Location,
				"address":        event.Address,
				"capacity":       event.Capacity,
				"photo":          event.Photo,
				"price":          event.Price,
				"status":         event.Status,
				"average_rating": averageRating,
				"unique_raters":  uniqueRaters,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"username":            loggedInUser.Username,
		"unregistered_events": events,
	})
}


