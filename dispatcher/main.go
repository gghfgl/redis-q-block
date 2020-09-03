package main

import (
	"fmt"
	//"time"
	"os"
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// NOTE: Handler
type RedisConfig struct {
	client *redis.Client
	qKey   string
}

type Handler struct {
	logger echo.Logger
	rdc    RedisConfig
}

func NewHandler(rdc RedisConfig) *Handler {
	e := echo.New()
	logger := e.Logger
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.INFO)
	logger.SetHeader("time=${time_rfc3339}, level=${level}, message=${message}")

	return &Handler{
		logger: logger,
		rdc:    rdc,
	}
}

func (h *Handler) RPushHandler(c echo.Context) error {
	data := c.FormValue("data")
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"status":  "400",
			"message": "Cannot submit empty field" ,
		})
	}

	err := h.rdc.client.RPush(h.rdc.qKey, data).Err()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"status":  "400",
			"message": err.Error(),
		})
	}

	// DEBUG
	h.logger.Infof("-> rpush to \"%s\" success ...", qKey)
	
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "200",
		"message": "success",
	})
}

// NOTE: Main
const qKey = "blockQueue"

func main() {
	// Redis
	rdc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := rdc.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ping redis ok!")

	// Handler
	h := NewHandler(RedisConfig{
		client: rdc,
		qKey: qKey,
	})

	// Http server
	e := echo.New()
	e.POST("/rpush", h.RPushHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
