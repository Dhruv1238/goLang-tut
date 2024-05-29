package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
)

type Rectangles []Rectangle

type Circles []Circle

type Rectangle struct {
	Length float32 `json:"length"`
	Width  float32 `json:"breadth"`
	Color  string  `json:"color"`
	Area   float32 `json:"area, omitempty"`
}

type Circle struct {
	Radius float32 `json:"radius"`
	Color  string  `json:"color"`
	Area   float32 `json:"area, omitempty"`
}

type Json struct {
	Rectangle Rectangles `json:"rectangles"`
	Circle    Circles    `json:"circles"`
}

func SortByArea(jsonData Json) Json {
	// var rectangles []Rectangles
	// var circles []Circles

	for i, shape := range jsonData.Rectangle {
		area := shape.Length * shape.Width
		jsonData.Rectangle[i].Area = area
	}

	for i, shape := range jsonData.Circle {
		area := 3.14 * shape.Radius * shape.Radius
		jsonData.Circle[i].Area = area
	}

	sort.Slice(jsonData.Rectangle, func(i, j int) bool {
		return jsonData.Rectangle[i].Area < jsonData.Rectangle[j].Area
	})

	sort.Slice(jsonData.Circle, func(i, j int) bool {
		return jsonData.Circle[i].Area < jsonData.Circle[j].Area
	})

	return Json{jsonData.Rectangle, jsonData.Circle}
}

func PrintJson(jsonData Json) {
	fmt.Println("Rectangles:")
	for _, shape := range jsonData.Rectangle {
		fmt.Println("{' Length ':", shape.Length, "' Width ':", shape.Width, "' Color ':", "'", shape.Color, "'", "}")
	}

	fmt.Println("Circles:")
	for _, shape := range jsonData.Circle {
		fmt.Println("{' Radius ':", shape.Radius, "' Color ':", "'", shape.Color, "'", "}")
	}
}

func main() {
	jsonFlag := flag.String("json", "", "Print JSON output")
	jsonFileFlag := flag.String("json-file", "", "Print JSON output from file")
	portFlag := flag.String("port", "", "Port to run the server on")

	flag.Parse()

	if *jsonFlag != "" {

		var jsonData Json
		err := json.Unmarshal([]byte(*jsonFlag), &jsonData)
		if err != nil {
			fmt.Println("Error While Processing Json", err)
		}

		jsonData = SortByArea(jsonData)

		PrintJson(jsonData)

		// fmt.Println(jsonData)
	} else if *jsonFileFlag != "" {
		data, err := ioutil.ReadFile(*jsonFileFlag)
		if err != nil {
			fmt.Println("Error While Reading File", err)
		}
		var jsonData Json
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			fmt.Println("Error While Processing Json", err)
		}
		jsonData = SortByArea(jsonData)

		PrintJson(jsonData)
	} else if *portFlag != "" {
		router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/shapes", func(w http.ResponseWriter, r *http.Request) {
			var jsonData Json
			err := json.NewDecoder(r.Body).Decode(&jsonData)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			jsonData = SortByArea(jsonData)
			json.NewEncoder(w).Encode(jsonData)
		}).Methods("POST")

		http.ListenAndServe(":"+*portFlag, router)
	} else {
		fmt.Println("Invalid Arguments")
	}

}
