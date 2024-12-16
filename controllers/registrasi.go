package controllers

import (
	"backend-event/database"
	"backend-event/models"
	"net/http"
	"log"
	"fmt"
	"os"
	"strings"

	gomail "gopkg.in/mail.v2"

	"github.com/gin-gonic/gin"
)

func RegisterEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	var event models.Event

	if err := database.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event tidak ditemukan"})
		return
	}

	if event.RemainingCapacity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event sudah penuh"})
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

	var userEvent models.Registration
	if err := database.DB.Where("user_id = ? AND event_id = ?", loggedInUser.ID, event.ID).First(&userEvent).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah terdaftar untuk event ini"})
		return
	}

	var input struct {
		Name          string `json:"name"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		Job           string `json:"job"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if strings.ToLower(event.Price) != "free" && input.PaymentMethod == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Metode pembayaran diperlukan untuk event berbayar"})
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userEvent = models.Registration{
		UserID:        loggedInUser.ID,
		EventID:       event.ID,
		Name:          input.Name,
		Username:      loggedInUser.Username,
		Email:         input.Email,
		PhoneNumber:   input.Phone,
		Job:           input.Job,
		PaymentMethod: input.PaymentMethod,
	}

	if err := tx.Create(&userEvent).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftar event"})
		return
	}

	event.RemainingCapacity -= 1
	if err := tx.Save(&event).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kapasitas event"})
		return
	}

	if strings.ToLower(event.Price) != "free" && input.PaymentMethod != "" {
		if err := sendPaymentConfirmationEmail(userEvent.Email, userEvent.Name, event.Name, event.Description, event.DateStart, event.Location, input.PaymentMethod); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim email konfirmasi pembayaran"})
			return
		}	
	}

	if err := sendEmail(userEvent.Email, userEvent.Name, userEvent.PhoneNumber, userEvent.Job, event.Name, event.Location, event.DateStart, event.Description, event.Mode, event.Link, event.Address); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim email konfirmasi pendaftaran"})
		return
	}
	

	tx.Commit()

	if err := UpdatePopularityScore(event.ID); err != nil {
		log.Printf("Gagal memperbarui popularity score: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mendaftar untuk event"})
}


func sendPaymentConfirmationEmail(to, name, eventName, description, eventDate, eventLocation, paymentMethod string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Konfirmasi Pembayaran - "+eventName)

	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Halo, %s!</h2>
			<p>Terima kasih telah melakukan pembayaran untuk event:</p>
			
			<div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 15px 0;">
				<h3 style="color: #007bff; margin-top: 0;">%s</h3>
				<p style="color: #666;"><strong>Deskripsi:</strong> %s</p>
				<p style="color: #666;"><strong>Tanggal:</strong> %s</p>
				<p style="color: #666;"><strong>Lokasi:</strong> %s</p>
				<p style="color: #666;"><strong>Metode Pembayaran:</strong> %s</p>
			</div>

			<p>Pembayaran Anda telah berhasil diproses. Silakan simpan email ini sebagai bukti pembayaran.</p>
			
			<div style="margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee;">
				<p style="color: #666; font-size: 14px;">
					Jika Anda memiliki pertanyaan, silakan hubungi tim support kami di:<br>
					Email: anjarriho081@gmail.com<br>
					WhatsApp: +62 890 3333 4444
				</p>
			</div>
		</div>
	`, name, eventName, description, eventDate, eventLocation, paymentMethod)

	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func sendEmail(to, name, phone, job, eventName, eventLocation, eventDate, description, mode, link, address string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Konfirmasi Pendaftaran - "+eventName)

	var locationTemplate string
	if mode == "online" {
		locationTemplate = fmt.Sprintf(`
			<div style="background-color: #e3f2fd; padding: 15px; border-radius: 5px; margin: 15px 0;">
				<h4 style="color: #0d47a1; margin-top: 0;">Informasi Meeting:</h4>
				<p style="color: #666;"><strong>Platform:</strong> %s</p>
				<p style="color: #666;"><strong>Link:</strong> 
					<a href="%s" style="color: #007bff; text-decoration: none;">Klik di sini untuk bergabung</a>
				</p>
				<p style="color: #856404; font-size: 14px;">
					<strong>Catatan:</strong> Pastikan Anda memiliki aplikasi yang diperlukan dan koneksi internet yang stabil
				</p>
			</div>
		`, eventLocation, link)
	} else {
		locationTemplate = fmt.Sprintf(`
			<div style="background-color: #f8f9fa; padding: 15px; border-radius: 5px; margin: 15px 0;">
				<h4 style="color: #28a745; margin-top: 0;">Lokasi Event:</h4>
				<p style="color: #666;"><strong>Lokasi:</strong> %s</p>
				<p style="color: #666;"><strong>Alamat Lengkap:</strong> %s</p>
			</div>
		`, eventLocation, address)
	}

	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Halo, %s!</h2>
			<p>Terima kasih telah mendaftar pada event kami:</p>
			
			<div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 15px 0;">
				<h3 style="color: #007bff; margin-top: 0;">%s</h3>
				<p style="color: #666;"><strong>Deskripsi:</strong> %s</p>
				<p style="color: #666;"><strong>Tanggal:</strong> %s</p>
				<p style="color: #666;"><strong>Mode Event:</strong> %s</p>
			</div>

			%s

			<div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 15px 0;">
				<h4 style="color: #28a745; margin-top: 0;">Detail Pendaftar:</h4>
				<p style="color: #666;"><strong>Nama:</strong> %s</p>
				<p style="color: #666;"><strong>Telepon:</strong> %s</p>
				<p style="color: #666;"><strong>Pekerjaan:</strong> %s</p>
			</div>

			<div style="background-color: #fff3cd; padding: 15px; border-radius: 5px; margin: 15px 0;">
				<p style="color: #856404; margin: 0;">
					<strong>Penting:</strong>
				</p>
				<ul style="color: #856404; margin: 5px 0;">
					%s
				</ul>
			</div>

			<div style="margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee;">
				<p style="color: #666; font-size: 14px;">
					Untuk informasi lebih lanjut sudah tertera diwebsite, atau silakan hubungi:<br>
					Email: anjarriho081@gmail.com<br>
					WhatsApp: +62 890 3333 4444
				</p>
			</div>
		</div>
	`, 
	name, eventName, description, eventDate, mode, 
	locationTemplate,
	name, phone, job,
	getImportantNotes(mode))

	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func getImportantNotes(mode string) string {
	if mode == "online" {
		return `
			<li>Bergabung ke meeting 10 menit sebelum acara dimulai</li>
			<li>Pastikan koneksi internet Anda stabil</li>
			<li>Siapkan mikrofon dan kamera jika diperlukan</li>
			<li>Gunakan headphone untuk kualitas audio yang lebih baik</li>
		`
	}
	return `
		<li>Harap datang 30 menit sebelum acara dimulai</li>
		<li>Membawa kartu identitas</li>
		<li>Berpakaian rapi dan sopan</li>
	`
}
