package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"net/http"
	"os"
	"strings"

	"github.com/aeperfilev/instafix/config"
	"github.com/aeperfilev/instafix/pkg/instafix"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func main() {
	var (
		configPath string
		addr       string
	)
	flag.StringVar(&configPath, "config", "", "Path to profiles.toml (optional)")
	flag.StringVar(&addr, "addr", "", "Listen address (defaults to :8080 or :$PORT)")
	flag.Parse()

	if strings.TrimSpace(addr) == "" {
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			addr = ":8080"
		} else if strings.HasPrefix(port, ":") {
			addr = port
		} else {
			addr = ":" + port
		}
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}

	processor, err := instafix.NewProcessor(cfg)
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/fix", authMiddleware(), func(c *gin.Context) {
		handleFix(c, processor)
	})

	if err := router.Run(addr); err != nil {
		panic(err)
	}
}

func handleFix(c *gin.Context, processor *instafix.Processor) {
	profileName := strings.TrimSpace(c.DefaultQuery("profile", "default"))
	watermark := c.Query("watermark")

	var srcImg image.Image
	var err error

	if fileHeader, errMultipart := c.FormFile("image"); errMultipart == nil {
		file, errOpen := fileHeader.Open()
		if errOpen != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read image file"})
			return
		}
		defer file.Close()
		srcImg, err = instafix.DecodeImage(file, fileHeader.Filename)
	} else {
		if c.Request.Body == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty request body"})
			return
		}
		defer c.Request.Body.Close()
		srcImg, err = instafix.DecodeImage(c.Request.Body, "")
	}
	if err != nil {
		logRequestError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image format or decode failed: " + err.Error()})
		return
	}

	result, quality, err := processor.Process(srcImg, profileName, watermark)
	if err != nil {
		status := http.StatusBadRequest
		if !isUserError(err) {
			status = http.StatusInternalServerError
		}
		logRequestError(c, err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "image/jpeg")
	if err := imaging.Encode(c.Writer, result, imaging.JPEG, imaging.JPEGQuality(quality)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encode failed"})
	}
}

func loadConfig(path string) (config.Config, error) {
	if strings.TrimSpace(path) == "" {
		cfg, _, err := config.LoadDefault()
		return cfg, err
	}
	return config.Load(path)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(os.Getenv("API_KEY"))
		if apiKey == "" {
			c.Next()
			return
		}
		if c.GetHeader("X-API-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func isUserError(err error) bool {
	var userErr instafix.UserError
	return errors.As(err, &userErr)
}

func logRequestError(c *gin.Context, err error) {
	contentType := c.GetHeader("Content-Type")
	profile := c.DefaultQuery("profile", "default")
	watermark := c.Query("watermark")
	bodyLen := 0
	if c.Request != nil && c.Request.ContentLength > 0 {
		bodyLen = int(c.Request.ContentLength)
	}
	c.Error(err)
	gin.DefaultErrorWriter.Write([]byte(
		fmt.Sprintf("fix error: %v | content_type=%s profile=%s watermark=%q body_len=%d\n",
			err, contentType, profile, watermark, bodyLen),
	))
}
