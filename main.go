package main

import (
    "fmt"

    "github.com/mgordon34/kornet-kover/pkg/scraper"
    "github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    fmt.Println("Hello there") 
    scraper.Scrape("test")

    storage.InitDB()
    storage.InitTables()
    _ = storage.GetDB()
}
