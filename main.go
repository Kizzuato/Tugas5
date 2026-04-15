package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Inisialisasi SQLite
	db, err := sql.Open("sqlite3", "./bmi.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Buat tabel jika belum ada
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, berat REAL, tinggi REAL, bmi REAL)")
	statement.Exec()

	r := gin.Default()

	// --- 1. MIDDLEWARE CORS (PENTING!) ---
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// --- 2. SERVE HTML (Agar tidak 404 saat akses localhost:8080) ---
	// Pastikan file index.html ada di folder yang sama dengan Dockerfile
	r.StaticFile("/", "./index.html")

	// Endpoint POST untuk hitung BMI
	r.POST("/data", func(c *gin.Context) {
		var input struct {
			Berat  float64 `json:"berat"`
			Tinggi float64 `json:"tinggi"`
		}
		
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Input tidak valid"})
			return
		}

		if input.Tinggi <= 0 {
			c.JSON(400, gin.H{"error": "Tinggi harus lebih dari 0"})
			return
		}

		tinggiMeter := input.Tinggi / 100
		bmi := input.Berat / (tinggiMeter * tinggiMeter)

		ins, err := db.Prepare("INSERT INTO history (berat, tinggi, bmi) VALUES (?, ?, ?)")
		if err != nil {
			c.JSON(500, gin.H{"error": "Gagal menyimpan data"})
			return
		}
		ins.Exec(input.Berat, input.Tinggi, bmi)

		c.JSON(200, gin.H{
			"berat":  input.Berat,
			"tinggi": input.Tinggi,
			"bmi":    fmt.Sprintf("%.2f", bmi),
		})
	})

	// Endpoint GET untuk mengambil semua data history
	r.GET("/data", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, berat, tinggi, bmi FROM history ORDER BY id DESC")
		if err != nil {
			c.JSON(500, gin.H{"error": "Gagal mengambil data"})
			return
		}
		defer rows.Close()

		var history []map[string]interface{}
		for rows.Next() {
			var id int
			var berat, tinggi, bmi float64
			rows.Scan(&id, &berat, &tinggi, &bmi)
			history = append(history, map[string]interface{}{
				"id":     id,
				"berat":  berat,
				"tinggi": tinggi,
				"bmi":    fmt.Sprintf("%.2f", bmi),
			})
		}

		c.JSON(200, history)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on port " + port)
	r.Run(":" + port)
}