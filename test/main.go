package main

import (
	"log"

	"github.com/hosting-de-labs/go-confluence-server-api"
)

func main() {
	api, err := confluence.NewAPI("https://hostname/confluence/rest/api", "user", "pass")
	if err != nil {
		log.Fatal(err)
	}

	//parentId := "64819202"
	id := "64819209"

	page, err := api.GetPage(id)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	/*str, err := json.Marshal(page)

	fmt.Printf("OOO %s\n", str)

	return*/

	body := page.Body.Storage.Value
	body += "<p>2. Test</p>"

	version := page.Version.Number + 1

	page, err = api.UpdatePage("ISMS", id, "monitor-0026", body, version)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	//fmt.Printf("%s", page)
}
