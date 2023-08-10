// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package Main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type ResponseData struct {
	Subdomain  string `json:"subdomain"`
	Response   string `json:"response"`
	StatusCode int    `json:"statusCode"`
}

var addr = flag.String("addr", ":8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func runTheTool(c *websocket.Conn, mt int, message []byte, ctx context.Context) {

	file, err := os.Open("wordlists.txt")
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}
	defer file.Close()

	// Create scanner
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			fmt.Println("Subdomain validation cancelled")
			return
		default:
			line := scanner.Text()
			subdomain := line + "." + string(message)

			url := "http://" + subdomain

			resp, err := net.LookupHost(subdomain)

			if err != nil {

			} else {
				resp2, err := http.Get(url)
				if err != nil {

				} else {
					data := ResponseData{
						Subdomain:  subdomain,
						Response:   strings.Join(resp, ", "),
						StatusCode: resp2.StatusCode,
					}

					// Encode the struct to JSON
					jsonData, err := json.Marshal(data)
					if err != nil {
						log.Println("json encode err:", err)
						break
					}
					c.WriteMessage(mt, jsonData)
					fmt.Println(subdomain, resp, resp2.StatusCode)
					resp2.Body.Close()
				}
			}
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fmt.Println(message)
		if string(message) == "stop-evaluating" {
			c.Close()
			cancel()
			fmt.Println("Stopping")
		}

		runTheTool(c, mt, message, ctx)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("main.html")
	if err != nil {
		panic(err)
	}

	err = t.Execute(w, "ws://"+r.Host+"/echo")
	if err != nil {
		panic(err)
	}
}

func Main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	// mux := http.NewServeMux()
	// mux.Handle("/static/", twhandler.New(http.Dir("static"), "static", twembed.New()))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
