package datastore

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Controller struct {
	Repository Repository
}

func (c *Controller) getPillsWithMicroNutrients(w http.ResponseWriter, r *http.Request) {
	var micronutrients Micronutrients
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576)) // read the body of the request
	if err != nil {
		log.Fatalln("Error getPillsWithMicroNutrients", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error getPillsWithMicroNutrients", err)
	}

	if err := json.Unmarshal(body, &micronutrients); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		log.Println(err)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error MicroNutrients unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	inventorychan := make(chan *PillInventories)
	go getPillsForMicronutrients(inventorychan, &micronutrients)
	result := <-inventorychan
	data, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
	return
}

func (c *Controller) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	var pillInventory PillInventory
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576)) // read the body of the request
	if err != nil {
		log.Fatalln("Error UpdateInventory", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error UpdateInventory", err)
	}

	if err := json.Unmarshal(body, &pillInventory); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		log.Println(err)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error pillInventory unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	//TODO

}

func (c *Controller) AddPill(w http.ResponseWriter, r *http.Request) {
	var pill Pill
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576)) // read the body of the request

	if err != nil {
		log.Fatalln("Error AddPill", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error AddPill", err)
	}

	if err := json.Unmarshal(body, &pill); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		log.Println(err)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error pill unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	respond := make(chan bool)
	go c.Repository.addPillAndUpdateInverseDictionary(respond, &pill)
	result := <-respond
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if result {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}
