package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/levigross/grequests"
)

const API_URL = "http://localhost/admin/api.php"
const BROADCAST_PORT = 32768
const HTTP_PORT = 15099

type AddAdlistRequest struct {
	Url string `json:"url"`
}

// type ChangeStateRequest struct {
// 	Enable  bool `json:"enable"`
// 	Timeout int  `json:"timeout"`
// }

type Adlist struct {
	Id      int    `json:"id"`
	Address string `json:"address"`
	Domains int    `json:"number"`
}

func sendRequest(context *gin.Context, requestParams map[string]string) {
	// Add auth if present
	var value, authExists = context.GetQuery("auth")
	if authExists {
		requestParams["auth"] = value
	}

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

		pc.WriteTo([]byte(fmt.Sprintf("AUTHCODE:%d", HTTP_PORT)), addr)
		pc.Close()
		fmt.Println(string(buf[:]))
		fmt.Println("Received broadcast from", addr)
	}
}

func main() {
	adminIp := make(chan string)
	go receiveBroadcast(adminIp)

	router := gin.Default()
	router.GET("/summary", func(context *gin.Context) {
		requestParams := map[string]string{
			"summaryRaw": "true",
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

	router.GET("/adlists", func(context *gin.Context) {
		requestParams := map[string]string{
			"list":        "adlist",
			"get_adlists": "true",
		}
		sendRequest(context, requestParams)
	})

	router.POST("/adlists", func(context *gin.Context) {
		// adlist is a form field with URL value
		log.Output(2, "Received adlist: "+context.PostForm("address"))
		body := AddAdlistRequest{}

		if err := context.BindJSON(&body); err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		log.Output(2, "Received adlist: "+body.Url)

		requestParams := map[string]string{
			"list":       "adlist",
			"add_adlist": body.Url,
		}
		sendRequest(context, requestParams)
	})

	router.DELETE("/adlists/:id", func(context *gin.Context) {
		requestParams := map[string]string{
			"list":          "adlist",
			"delete_adlist": context.Param("id"),
		}
		sendRequest(context, requestParams)
	})

	router.GET("/adlists/all", func(context *gin.Context) {
		var dataItems = []Adlist{
			{
				Id:      1,
				Address: "https://gitlab.science.ru.nl/bram/pihole-facebook/raw/master/pihole-facebook.txt",
				Domains: 10,
			},
			{
				Id:      2,
				Address: "https://github.com/Adlist2.txt",
				Domains: 20,
			},
		}

		var responseBody = make(map[string][]Adlist)
		responseBody["data"] = dataItems
		data, err := json.Marshal(responseBody)
		if err != nil {
			log.Fatal(err)
		}
		context.Data(http.StatusOK, gin.MIMEJSON, []byte(data))
	})

	// TODO: check this feature later.
	// router.GET("/blockables", func(context *gin.Context) {
	// 	requestParams := map[string]string{
	// 		"list": "blockables",
	// 	}
	// 	sendRequest(context, requestParams)
	// })

	// TODO: check this feature later.
	// whitelist = 0
	// blacklist = 1
	// regex_whitelist = 2
	// regex_blacklist = 3
	// router.POST("/blockables", func(context *gin.Context) {
	// 	requestParams := map[string]string{
	// 		"list": "blockables",
	// 		"add":  context.PostForm("value"),
	// 		"type": context.PostForm("type"),
	// 	}
	// 	sendRequest(context, requestParams)
	// })
	router.Run(fmt.Sprintf("0.0.0.0:%d", HTTP_PORT))
}
