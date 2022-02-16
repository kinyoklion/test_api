package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Echo struct {
	Body    string      `json:"body"`
	Headers http.Header `json:"headers"`
	Method  string      `json:"method"`
}

func chanWriter(opChan chan string) {
	for index := 0; index < 5; index++ {
		opChan <- fmt.Sprintf("num: %d", index)
		time.Sleep(time.Second)
	}
	close(opChan)
}

func stream(c *gin.Context) {
	opChan := make(chan string)
	go chanWriter(opChan)

	c.Stream(func(w io.Writer) bool {
		output, ok := <-opChan
		if !ok {
			return false
		}
		outputBytes := bytes.NewBufferString(output)
		c.Writer.Write(append(outputBytes.Bytes(), []byte("\r\n")...))
		return true
	})
}

func sse(c *gin.Context) {
	chanStream := make(chan int, 10)
	go func() {
		defer close(chanStream)
		for i := 0; i < 5; i++ {
			chanStream <- i
			time.Sleep(time.Second * 1)
		}
	}()
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-chanStream; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

func echo(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return
	}

	c.JSON(http.StatusOK, Echo{
		Body:    string(body),
		Headers: c.Request.Header,
		Method:  c.Request.Method,
	})
}

func main() {
	router := gin.Default()
	router.GET("/echo", echo)
	router.POST("/echo", echo)
	router.DELETE("/echo", echo)
	router.PATCH("/echo", echo)
	router.PUT("/echo", echo)

	router.GET("/stream", stream)

	router.GET("/sse", sse)

	router.Run("localhost:8080")
}
