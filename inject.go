package main

import (
  "bufio"
  "os"
  "regexp"
  "bytes"
  "encoding/json"
  "log"
  "fmt"
  "github.com/elastic/go-elasticsearch"
  "github.com/elastic/go-elasticsearch/esapi"
  "runtime"
  "strconv"
)

type Account struct {
  Username string `json:"username"`
  Password string `json:"password"`
}

var (
  res *esapi.Response
  es_bulk_buffer bytes.Buffer
  indexName = "none"
  bulk_insert_size = 10000
  re_extract, _ = regexp.Compile("(.*)[;:,](.*)")
  account_count = 0
)

func main() {
  fmt.Printf("Starting: account bulk importer\n")

  if len(os.Args) != 4 {
    log.Fatalf("Args error\n  Usage: %s <file to index> <index name> <bulk size>\n  Use env ELASTICSEARCH_URL to specify Elasticsearch dest\n", os.Args[0])
  }
  filename := os.Args[1]
  indexName := os.Args[2]
  bulk_insert_size, _ = strconv.Atoi(os.Args[3])

  // Open file containing accounts
  fmt.Printf("Opening file: %s\n", filename)
  file, err := os.Open(filename)
  if err != nil {
      log.Fatal(err)
  }
  defer file.Close()

  // Parsing file
  scanner := bufio.NewScanner(file)
  // For each line
  for scanner.Scan() {
    account_count++
    extract := re_extract.FindAllStringSubmatch(scanner.Text(), -1)
    account := &Account{Username: extract[0][1], Password: extract[0][2]}
    // Bulk header, which index do we want to use (for each line)
    meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" } }%s`, indexName, "\n"))
    // Converting struct into JSON
    json_data, err := json.Marshal(account)
    if err != nil {
      log.Fatalf("Cannot encode account: %s", err)
    }
    // Adding new line
    json_data = append(json_data, "\n"...)
    // Appending the account to the buffer
    es_bulk_buffer.Grow(len(meta) + len(json_data))
    es_bulk_buffer.Write(meta)
    es_bulk_buffer.Write(json_data)

    if account_count == bulk_insert_size {
      flushBufferToEs()
      // Cleaning the buffer and reset the counter
      es_bulk_buffer.Reset()
      account_count = 0
    }

  }
  // Flush remaining accounts
  if account_count != 0 {
    flushBufferToEs()
  }
  if err := scanner.Err(); err != nil {
      log.Fatal(err)
  }
}

func flushBufferToEs() {
  es, err := elasticsearch.NewDefaultClient()
  if err != nil {
    log.Fatalf("Error creating the client: %s", err)
  }

  res, err = es.Bulk(bytes.NewReader(es_bulk_buffer.Bytes()))
  if err != nil {
    log.Fatalf("Failure indexing batch: %s", err)
  }
  fmt.Printf("%d document(s) inserted (", account_count)
  PrintMemUsage()
  fmt.Printf(")\n")
}


// Thanks to https://golangcode.com/print-the-current-memory-usage/
func PrintMemUsage() {
  var m runtime.MemStats
  runtime.ReadMemStats(&m)
  // For info on each, see: https://golang.org/pkg/runtime/#MemStats
  fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
  fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
  fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
  fmt.Printf("\tNumGC = %v", m.NumGC)
}

func bToMb(b uint64) uint64 {
  return b / 1024 / 1024
}
