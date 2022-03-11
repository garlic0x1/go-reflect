package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func oracle(resp *http.Response, test string, f form, payload string) {
	respBytes, err := ioutil.ReadAll(resp.Body)
	respString := string(respBytes)
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(respString, "DDCCBBAA") {
		fmt.Println("Reflection found", f.Method, f.URL, payload)
	}
}

func testReflection(f form) {

	var client http.Client

	if f.Method == "POST" {
		params := url.Values{}
		for i := 0; i < len(f.Inputs); i++ {
			if f.Inputs[i].Type == "hidden" {
				params.Add(f.Inputs[i].Name, f.Inputs[i].Value)
			}
			if f.Inputs[i].Type == "email" {
				params.Add(f.Inputs[i].Name, `DDCCBBAA@gmail.com`)
			}
			if f.Inputs[i].Type == "text" {
				params.Add(f.Inputs[i].Name, `http://DDCCBBAA`)
			}
		}
		payload := strings.NewReader(params.Encode())
		req, err := http.NewRequest(f.Method, f.URL, payload)
		if err != nil {
			log.Println("error making request", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error performing request", err)
		}
		defer resp.Body.Close()

		oracle(resp, "DDCCBBAA", f, "")
	} else if f.Method == "GET" {
		payload := f.URL + "?"
		for i := 0; i < len(f.Inputs); i++ {
			if i != 0 {
				payload = payload + "&"
			}
			payload = payload + f.Inputs[i].Name + "=" + "DDCCBBAA"
		}
		req, err := http.NewRequest(f.Method, payload, nil)
		if err != nil {
			log.Println("error making request", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error performing request", err)
		}
		defer resp.Body.Close()
		oracle(resp, "DDCCBBAA", f, payload)
	}
}
