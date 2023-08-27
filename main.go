package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

const dbConn = "root@/test"

// Total connection idle
const dbMaxIdleConns = 4

// max connection pool
const dbMaxConns = 100

// total worker
const totalWorker = 100

// File csv
const csvFile = "majestic_million.csv"

var dataHeaders = make([]string, 0)

func main() {

	start := time.Now()

	db, err := openDbConnection()
	if err != nil {
		log.Fatalf(err.Error())
	}

	csvReader, csvFile, err := openCsvfile()
	if err != nil {
		log.Fatal(err.Error())
	}

	defer csvFile.Close()

	jobs := make(chan []interface{}, 0)
	wg := new(sync.WaitGroup)

	go dispatchWorkers(db, jobs, wg)
	readCsvFilePerLineThenSendToWorker(csvReader, jobs, wg)

	wg.Wait()

	duration := time.Since(start)
	fmt.Println("dine in", int(math.Ceil(duration.Seconds())), "seconds")
}

func openDbConnection() (*sql.DB, error) {
	log.Println("=> open db connection")

	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(dbMaxConns)
	db.SetMaxIdleConns(dbMaxIdleConns)

	return db, nil
}

func openCsvfile() (*csv.Reader, *os.File, error) {
	log.Println("=> open csv file")

	f, err := os.Open(csvFile)
	if err != nil {
		return nil, nil, err
	}

	reader := csv.NewReader(f)
	return reader, f, nil
}

func dispatchWorkers(db *sql.DB, jobs <-chan []interface{}, wg *sync.WaitGroup) {
	for workerIndex := 0; workerIndex <= totalWorker; workerIndex++ {
		go func(workerIndex int, db *sql.DB, jobs <-chan []interface{}, wg *sync.WaitGroup) {
			counter := 0

			for job := range jobs {
				doTheJob(workerIndex, counter, db, job)
				wg.Done()
				counter++
			}
		}(workerIndex, db, jobs, wg)
	}
}

func readCsvFilePerLineThenSendToWorker(csvReader *csv.Reader, jobs chan<- []interface{}, wg *sync.WaitGroup) {
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}

			break
		}

		// Insert row (0) / Header
		if len(dataHeaders) == 0 {
			dataHeaders = row
			continue
		}

		rowOrdered := make([]interface{}, 0)
		for _, each := range row {
			rowOrdered = append(rowOrdered, each)
		}

		wg.Add(1)
		// Send data to jobs to be processed inserted
		jobs <- rowOrdered
	}

	close(jobs)
}

func doTheJob(workerIndex, counter int, db *sql.DB, values []interface{}) {
	for {
		var outerError error
		// IIFE
		func(outerError *error) {
			defer func() {
				if err := recover(); err != nil {
					*outerError = fmt.Errorf("%v", err)
				}
			}()

			conn, err := db.Conn(context.Background())
			query := fmt.Sprintf("INSERT INTO domain (%s) VALUES (%s)",
				strings.Join(dataHeaders, ","),
				// Generate query insert (?)
				strings.Join(generateQuestionsMark(len(dataHeaders)), ","),
			)

			_, err = conn.ExecContext(context.Background(), query, values...)
			if err != nil {
				log.Fatal(err.Error())
			}

			err = conn.Close()
			if err != nil {
				log.Fatal(err.Error())
			}
		}(&outerError)

		if outerError == nil {
			break
		}
	}

	if counter%100 == 0 {
		log.Println("=> worker", workerIndex, "inserted", counter, "data")
	}
}

func generateQuestionsMark(n int) []string {
	s := make([]string, 0)
	for i := 0; i < n; i++ {
		s = append(s, "?")
	}
	return s
}
