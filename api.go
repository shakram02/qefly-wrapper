package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/levigross/grequests"
)

const API_URL = "http://localhost/admin/api.php"
const BROADCAST_PORT = 32768

func sendRequest(context *gin.Context, requestParams map[string]string) {
	resp, err := grequests.Get(API_URL, &grequests.RequestOptions{Params: requestParams})

	if err != nil {
		log.Fatal(err)
		context.Error(err)
	} else {
		context.Data(http.StatusOK, gin.MIMEJSON, resp.Bytes())
	}
}

func receiveBroadcast(c chan string) {
	for {
		pc, err := net.ListenPacket("udp4", fmt.Sprintf("%s:%d", "255.255.255.255", BROADCAST_PORT))

		if err != nil {
			log.Panic(err)
		}

		fmt.Println("Listening for broadcast on port", BROADCAST_PORT)
		fmt.Println("Waiting for admin connection...")

		var buf [32]byte
		_, addr, err := pc.ReadFrom(buf[:])

		if err != nil {
			log.Panic(err)
		}

		pc.WriteTo(buf[:], addr)
		pc.Close()
		fmt.Println(string(buf[:]))
		fmt.Println("Received broadcast from", addr)
	}
}
func main() {
	adminIp := make(chan string)
	go receiveBroadcast(adminIp)

	router := gin.Default()

	router.GET("/adlists", func(context *gin.Context) {
		requestParams := map[string]string{
			"list":        "adlist",
			"get_adlists": "true",
		}
		sendRequest(context, requestParams)
	})

	router.POST("/adlists", func(context *gin.Context) {
		requestParams := map[string]string{
			"list":       "adlist",
			"add_adlist": context.PostForm("adlist"),
		}
		sendRequest(context, requestParams)
	})

	router.GET("/blockables", func(context *gin.Context) {
		requestParams := map[string]string{
			"list": "blockables",
		}
		sendRequest(context, requestParams)
	})

	// whitelist = 0
	// blacklist = 1
	// regex_whitelist = 2
	// regex_blacklist = 3
	router.POST("/blockables", func(context *gin.Context) {
		requestParams := map[string]string{
			"list": "blockables",
			"add":  context.PostForm("value"),
			"type": context.PostForm("type"),
		}
		sendRequest(context, requestParams)
	})

	router.POST("/enable", func(context *gin.Context) {
		requestParams := map[string]string{
			"enable": "true",
		}
		sendRequest(context, requestParams)
	})

	router.POST("/disable", func(context *gin.Context) {
		requestParams := map[string]string{
			"disable": context.DefaultPostForm("timeout", "true"),
		}
		sendRequest(context, requestParams)
	})

	router.Run("0.0.0.0:5000")
}
