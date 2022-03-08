package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/levigross/grequests"
)

const API_URL = "http://localhost/admin/api.php"

func sendRequest(context *gin.Context, requestParams map[string]string) {
	resp, err := grequests.Get(API_URL, &grequests.RequestOptions{Params: requestParams})

	if err != nil {
		log.Fatal(err)
		context.Error(err)
	} else {
		context.Data(http.StatusOK, gin.MIMEJSON, resp.Bytes())
	}
}

func main() {

	// var API_URL string = os.Getenv("API_URL")

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

	router.Run("localhost:5000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
