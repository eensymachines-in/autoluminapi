package main

import (
	"bufio"
	"os"

	"github.com/eensymachines-in/authapi/handlers"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	// +++++++++++++++++++ reading the secrets into the environment
	file, err := os.Open("/run/secrets/auth_secrets")
	if err != nil {
		log.Errorf("Failed to read encryption secrets, please load those %s", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	// ++++++++++++++++++++++++++++++++ reading in the auth secret
	line, _, err := reader.ReadLine()
	if err != nil {
		log.Error("Error reading the auth secret from file")
	}
	os.Setenv("AUTH_SECRET", string(line))
	log.Infof("The authentication secret %s", os.Getenv("AUTH_SECRET"))
	// ++++++++++++++++++++ reading in the refresh secret
	line, _, err = reader.ReadLine()
	if err != nil {
		log.Error("Error reading the refr secret from file")
	}
	os.Setenv("REFR_SECRET", string(line))
	log.Infof("The refresh secret %s", os.Getenv("REFR_SECRET"))

}
func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Use(CORS)
	// devices group
	devices := r.Group("/devices")
	devices.Use(lclDbConnect())

	devices.POST("", handlers.HandlDevices)          // when creating new registrations
	devices.GET("/:serial", handlers.HandlDevices)   // when getting existing registrations
	devices.PATCH("/:serial", handlers.HandlDevices) // when modifying existing registration

	// Users group
	users := r.Group("/users")
	users.Use(lclDbConnect())

	users.POST("", handlers.HndlUsers)
	users.GET("/:email", handlers.HandlUser)
	users.PUT("/:email", tokenParse(), verifyUser(), handlers.HandlUser) // changing the user account details
	users.PATCH("/:email", b64UserCredsParse(), handlers.HandlUser)      // update password
	// +++++++++ to delete an account you need elevated permission and authentication token
	users.DELETE("/:email", tokenParse(), verifyRole(2), handlers.HandlUser)

	// will handle only authentication
	auths := r.Group("/authenticate")
	auths.Use(lclCacConnect()).Use(lclDbConnect()).Use(b64UserCredsParse())
	auths.POST("/:email", handlers.HandlAuth)

	// /authorize/?lvl=2
	// /authorize/?refresh=true
	authrz := r.Group("/authorize")
	authrz.Use(lclCacConnect()).Use(tokenParse())
	authrz.GET("", handlers.HndlAuthrz)
	authrz.DELETE("", handlers.HndlAuthrz)
	log.Fatal(r.Run(":8080"))
}
