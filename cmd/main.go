package main

import (
  "net/http"
  "os"

  "github.com/gin-gonic/gin"
)

func main() {
  r:= gin.Default()
  r.GET("/", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "hello world",
    })
  })

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  if err := r.Run(":" + port); err != nil {
    panic(err)
  }
}
