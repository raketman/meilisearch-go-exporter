package main

import (
	"github.com/meilisearch/meilisearch-go"
	"log"
	"sync"
)

func main() {
	var config Config

	config = Config{}

	config.Read("config.json")

	var client = meilisearch.NewClient(meilisearch.Config{
		Host: config.Host,
		APIKey: config.Key,
	})


	var baseGroup sync.WaitGroup
	baseGroup.Add(len(config.Works))

	// Для каждого запрос создадим индекс primary
	for _, itemWork := range config.Works {
		go func(work Work) {

			if work.DeleteBefore {
				_, deleteErr := client.Indexes().Delete(work.Index)

				if deleteErr != nil {
					log.Fatal("Delete index: ", deleteErr)
				}
			}

			// Create an index if your index does not already exist
			index, _ := client.Indexes().Get(work.Index)

			if index == nil {
				_, err := client.Indexes().Create(meilisearch.CreateIndexRequest{
					UID:        work.Index,
					PrimaryKey: work.Primary,
				})
				if err != nil {
					log.Fatal("Create index: ", err)
				}
			}

			if len (work.DisplayedAttributes) > 0 {


				_, err := client.Settings(work.Index).UpdateDisplayedAttributes(work.DisplayedAttributes)

				if err != nil {
					log.Fatal("Set DisplayedAttributes: ", err)
				}
			}
			if len (work.SearchableAttributes) > 0 {

				_, err := client.Settings(work.Index).UpdateSearchableAttributes(work.SearchableAttributes)

				if err != nil {
					log.Fatal("Set SearchableAttributes: ", err)
				}
			}


			var threadGroup sync.WaitGroup
			threadGroup.Add(work.Thread)

			for thread := 0; thread < work.Thread; thread++ {
				go func(thread int) {
					exporter := Exporter{
						Thread: thread,
					}

					exporter.Process(client, work)

					threadGroup.Done()
				}(thread)
			}

			threadGroup.Wait()
			// Отправим, что завершили
			baseGroup.Done()
		} (itemWork)
	}

	baseGroup.Wait()

	log.Println("READY")
}
