package datastore

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

type Repository struct{}

func getClient(respond chan<- *redis.Client) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	respond <- client
}

func getPillsForMicronutrients(respond chan<- *PillInventories, micronutrients *Micronutrients) {
	clientChannel := make(chan *redis.Client)
	go getClient(clientChannel)
	client := <-clientChannel
	pillIDs := make(map[string]int)
	//Let's just loop and get
	micronutrientsMap := make(map[string]int)
	for _, micronutrient := range *micronutrients {
		key := micronutrient.Name
		micronutrientsMap[key] = 1
		ids, err := client.HGet("MicroNutrients", key).Result()
		if err != nil {
			fmt.Println("Something failed")
		} else {
			pillList := []string{}
			err = json.Unmarshal([]byte(ids), &pillList)
			if err != nil {
				fmt.Println("Failed to get list of ids")
			} else {
				for _, id := range pillList {
					pillIDs[id] = 1
				}
			}
		}
	}
	var inventories PillInventories
	for k, _ := range pillIDs {
		pillData, err := client.HGet("Pills", k).Result()
		if err != nil {
			continue
		}
		var pill Pill
		err = json.Unmarshal([]byte(pillData), &pill)
		if err != nil {
			continue
		}
		doesNotMatchMore := true
		for _, pillMicronutrient := range pill.PillMicronutrients {
			if _, ok := micronutrientsMap[pillMicronutrient.MicroNutrient.Name]; !ok {
				doesNotMatchMore = false
				continue
			}
		}
		if !doesNotMatchMore {
			continue
		}
		pillInventoryData, err := client.HGet("PillsInventory", k).Result()
		if err != nil {
			continue
		}
		inventory, err := strconv.Atoi(pillInventoryData)
		if err != nil {
			continue
		}
		pillInventory := PillInventory{PillData: &pill, Inventory: inventory}
		inventories = append(inventories, pillInventory)
	}
	respond <- &inventories
}

func addMicroNutrient(respond chan<- bool, microNutrient *Micronutrient) {
	clientChannel := make(chan *redis.Client)
	go getClient(clientChannel)
	client := <-clientChannel
	key := microNutrient.Name
	exists, err := client.HExists("MicroNutrients", key).Result()
	if err != nil {
		respond <- false
	}
	if exists {
		respond <- true
	}
	emptyList := []string{}
	emptyListData, err := json.Marshal(emptyList)
	err = client.HSet("MicroNutrients", key, emptyListData).Err()
	if err != nil {
		fmt.Println("Failed to add MicroNutrient")
		fmt.Println(err)
		respond <- false
	}
	respond <- true
}

func addToInverseDictionary(respond chan<- bool, pill *Pill, microNutrient *Micronutrient) {
	clientChannel := make(chan *redis.Client)
	go getClient(clientChannel)
	client := <-clientChannel
	key := microNutrient.Name
	exists, err := client.HExists("MicroNutrients", key).Result()
	if err != nil {
		respond <- false
	}
	if !exists {
		nutrientChannel := make(chan bool)
		fmt.Println("Adding MicroNutrient: " + key)
		go addMicroNutrient(nutrientChannel, microNutrient)
		result := <-nutrientChannel
		if !result {
			fmt.Println("Failed...")
			respond <- false
		}
	}
	keyPill := pill.Name
	pillList := []string{}
	pillListStr, err := client.HGet("MicroNutrients", key).Result()
	if err != nil {
		fmt.Println("pillListStr HGET fail")
		respond <- false
	}
	err = json.Unmarshal([]byte(pillListStr), &pillList)
	if err != nil {
		respond <- false
	}
	pillList = append(pillList, string(keyPill))
	for _, element := range pillList {
		fmt.Println(element)
	}
	pillListBytes, err := json.Marshal(pillList)
	if err != nil {
		respond <- false
	}
	err = client.HSet("MicroNutrients", key, string(pillListBytes)).Err()
	if err != nil {
		respond <- false
	}
	respond <- true
}

func (r Repository) addPillAndUpdateInverseDictionary(respond chan<- bool, pill *Pill) {
	clientChannel := make(chan *redis.Client)
	go getClient(clientChannel)
	client := <-clientChannel
	key := pill.Name
	//We should get a write lock here
	exists, err := client.HExists("Pills", key).Result()
	//Return if this pill has been registered
	if err != nil || exists {
		respond <- false
	}
	pilldata, err := json.Marshal(pill)
	if err != nil {
		respond <- false
	}
	err = client.HSet("Pills", key, pilldata).Err()
	if err != nil {
		fmt.Println(err)
		//We should do something
		respond <- false
	}
	//For simplicity create with 100 as inventory
	err = client.HSet("PillsInventory", key, 100).Err()
	if err != nil {
		fmt.Println(err)
		//We should do something
		respond <- false
	}
	pillMicronutrients := pill.PillMicronutrients
	numMicroNutrients := len(pillMicronutrients)
	dictChannel := make(chan bool, numMicroNutrients)
	for _, pillMicronutrient := range pillMicronutrients {
		go addToInverseDictionary(dictChannel, pill, pillMicronutrient.MicroNutrient)
	}
	var result = true
	for i := 0; i < numMicroNutrients; i++ {
		temp := <-dictChannel
		if !temp {
			result = false
		}
	}
	if result {
		respond <- true
	} else {
		respond <- false
	}
}
