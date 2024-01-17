package commands

// getCompliment returns a compliment for the given target
//func getCompliment(target string) {
//	log.Println("Getting compliment for " + target)
//
//	// Send GET request to API
//	resp, err := http.Get("https://complimentr.com/api")
//	if err != nil {
//		log.Printf("Error while calling the Compliment API: %v", err)
//	}
//
//	defer func(Body io.ReadCloser) {
//		_ = Body.Close()
//	}(resp.Body)
//
//	// Parse response JSON
//	var data map[string]interface{}
//
//	err = json.NewDecoder(resp.Body).Decode(&data)
//	if err != nil {
//		log.Printf("Error while decoding the Compliment API response: %v", err)
//	}
//
//	// Extract insult from response data
//	compliment, ok := data["compliment"].(string)
//	if !ok {
//		log.Println("Error while parsing the Compliment API response")
//	}
//
//	time.Sleep(1000 * time.Millisecond)
//
//	network.RconExecute("say \"" + target + " " + compliment + "\"")
//	log.Println("Compliment: " + compliment)
//}
