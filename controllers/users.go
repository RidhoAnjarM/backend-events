package controllers

import (
	"backend-event/models"
	"net/http"
	"backend-event/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)


func GetUserById(c *gin.Context) {
    var user models.User
    id := c.Param("id")

    if err := database.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id": user.ID,
        "username": user.Username,
    })
}

func GetAllUsers(c *gin.Context) {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve users",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users retrieved successfully",
		"users":   users,
	})
}

func UpdateUser(c *gin.Context) {
    id := c.Param("id")
    var user models.User

    if err := database.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error":   "User not found",
            "message": err.Error(),
        })
        return
    }

    var input struct {
        Username    string `json:"username"`
        Password    string `json:"password"`    
        NewPassword string `json:"newPassword"` 
        Role        string `json:"role"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid input",
            "message": err.Error(),
        })
        return
    }

    if input.Username != "" {
        user.Username = input.Username
    }

    if input.NewPassword != "" {
        if input.Password == "" {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Password lama harus diisi untuk verifikasi",
            })
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Password lama tidak sesuai",
            })
            return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Gagal mengenkripsi password",
            })
            return
        }

        user.Password = string(hashedPassword)
    }

    if input.Role != "" {
        user.Role = input.Role
    }

    if err := database.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to update user",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "User updated successfully",
        "user": gin.H{
            "id":       user.ID,
            "username": user.Username,
            "role":     user.Role,
        },
    })
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"message": err.Error(),
		})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user":    user,
	})
}
