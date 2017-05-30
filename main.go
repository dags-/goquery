package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"bufio"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/dags-/goquery/discord"
	"github.com/dags-/goquery/minecraft"
	"github.com/qiangxue/fasthttp-routing"
)

type Response struct {
	Result string `json:"result"`
	Time   string `json:"time"`
	Data   interface{} `json:"data"`
}

func main() {
	go handleStop()

	var port string
	flag.StringVar(&port, "port", "8080", "Query port")
	flag.Parse()

	notFound, readErr := ioutil.ReadFile("notfound.html")

	router := routing.New()

	router.Get("/discord/<id>", discordHandler)
	router.Get("/minecraft/<ip>", minecraftHandler)
	router.Get("/minecraft/<ip>/<port>", minecraftHandler)

	if readErr == nil {
		router.NotFound(func(c *routing.Context) error {
			c.Response.Header.SetContentType("text/html")
			_, err := c.Response.BodyWriter().Write(notFound)
			return err
		})
	} else {
		fmt.Println(readErr)
	}

	fmt.Println("Launching on port", port)
	panic(fasthttp.ListenAndServe(":8080", router.HandleRequest))
}


func handleStop() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			fmt.Println("Stopping...")
			os.Exit(0)
			break
		}
	}
}

func minecraftHandler(c *routing.Context) error {
	ip := c.Param("ip")
	port := c.Param("port")

	if port == "" {
		port = "25565"
	}

	status, err := goquery.GetStatus(ip, port)
	response := wrapResponse(status, err)

	return writeResponse(response, c)
}

func discordHandler(c *routing.Context) error {
	id := c.Param("id")
	data, err := discord.GetStatus(id)
	response := wrapResponse(data, err)
	return writeResponse(response, c)
}

func wrapResponse(data interface{}, err error) Response {
	var result = fmt.Sprint(err)
	var timestamp = time.Now().Format(time.RFC822)

	if err == nil {
		result = "success"
	}

	return Response{Result: result, Time: timestamp, Data: data}
}

func writeResponse(resp Response, c *routing.Context) error {
	var prefix, indent = "", ""
	if string(c.FormValue("pretty")) == "true" {
		indent = "  "
	}

	c.Response.Header.SetStatusCode(http.StatusOK)
	c.Response.Header.Set("Cache-Control", "max-age=60")
	c.Response.Header.Set("Access-Control-Allow-Origin", "*")
	c.Response.Header.SetContentType("application/json; charset=UTF-8")

	encoder := json.NewEncoder(c.Response.BodyWriter())
	encoder.SetIndent(prefix, indent)

	return encoder.Encode(resp)
}