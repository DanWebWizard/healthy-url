package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"healthy-url/cassandra"
	"healthy-url/models"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var workerCount int

func main() {
	// Load environment variables
	if err := loadEnv(); err != nil {
		fmt.Printf("Error loading environment variables: %v\n", err)
		return
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	logger.Infof("Initializing Cassandra...")
	InitCassandra(logger)
	defer cassandra.CloseCassandra()

	logger.Infof("Starting to check healthy urls!")
	urlHealths := CheckHealthOfUrls(logger)

	for i, urlHealth := range urlHealths {
		logger.Infof("URL %d: %s, Healthy: %t, Unhealthy: %t, Unreachable: %t, Time taken: %d ms, Status code: %d",
			i+1, urlHealth.URL, urlHealth.HealthyUrl, urlHealth.UnhealthyUrl, urlHealth.UnreachableUrl, urlHealth.TimeTaken, urlHealth.StatusCode)
	}
	logger.Info("Program exiting...")
}

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	workerCountStr := os.Getenv("WORKERS")
	var err error
	workerCount, err = strconv.Atoi(workerCountStr)
	if err != nil {
		workerCount = 5 // Default to 5 workers if not set
	}

	return nil
}

func CheckHealthOfUrls(logger *logrus.Logger) []models.UrlHealth {
	file, err := os.Open("urls.json")
	if err != nil {
		logger.Errorf("Error reading the file: %v", err)
		return nil
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("Error opening file: %v", err)
		return nil
	}

	var urls []string
	json.Unmarshal(byteValue, &urls)

	overallStartTime := time.Now()

	var urlHealths []models.UrlHealth
	var wg sync.WaitGroup
	urlChan := make(chan string, len(urls))
	resultChan := make(chan models.UrlHealth, len(urls))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(logger, urlChan, resultChan, &wg)
	}

	for _, url := range urls {
		urlChan <- url
	}

	close(urlChan)
	wg.Wait()
	close(resultChan)

	for urlHealth := range resultChan {
		urlHealths = append(urlHealths, urlHealth)
	}

	overallTimeTaken := int(time.Since(overallStartTime).Seconds())

	logger.Infof("Total time taken to check the health of all the urls: %d seconds", overallTimeTaken)

	return urlHealths
}

func worker(logger *logrus.Logger, urlChan chan string, resultChan chan models.UrlHealth, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range urlChan {
		startTime := time.Now()
		status, statusCode := GetHealth(url)
		timeTaken := int(time.Since(startTime).Milliseconds())

		urlHealth := models.UrlHealth{
			URL:        url,
			TimeTaken:  timeTaken,
			StatusCode: statusCode,
		}

		switch status {
		case models.Healthy:
			urlHealth.HealthyUrl = true
		case models.Unhealthy:
			urlHealth.UnhealthyUrl = true
		case models.Unreachable:
			urlHealth.UnreachableUrl = true
		}

		resultChan <- urlHealth

		if err := WriteUrlToCassandra(logger, urlHealth); err != nil {
			logger.Errorf("Error writing to Cassandra: %v", err)
		}
	}
}

func GetHealth(url string) (models.URLStatus, int) {
	resp, err := http.Get(url)
	if err != nil {
		logrus.Errorf("Error reaching the URL: %s, %v", url, err)
		return models.Unreachable, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return models.Healthy, resp.StatusCode
	}
	return models.Unhealthy, resp.StatusCode
}

func WriteUrlToCassandra(logger *logrus.Logger, urlHealth models.UrlHealth) error {
	logger.Debugf("Writing url health to Cassandra: URL=%s, HealthyUrl=%t, UnhealthyUrl=%t, UnreachableUrl=%t, TimeTaken=%d, StatusCode=%d",
		urlHealth.URL, urlHealth.HealthyUrl, urlHealth.UnhealthyUrl, urlHealth.UnreachableUrl, urlHealth.TimeTaken, urlHealth.StatusCode)
	if err := cassandra.InsertUrlHealth(logger, urlHealth); err != nil {
		return fmt.Errorf("failed to write url health to Cassandra: %v", err)
	}
	logger.Debug("Successfully wrote url health to Cassandra")
	return nil
}

func InitCassandra(logger *logrus.Logger) {
	// Initialize Cassandra
	cassandraHost := os.Getenv("CASSANDRA_HOST")
	if cassandraHost == "" {
		cassandraHost = "cassandra" // Default to the service name defined in docker-compose.yml
	}
	cassandraHosts := []string{cassandraHost}
	keyspace := os.Getenv("CASSANDRA_KEYSPACE")
	if keyspace == "" {
		keyspace = "web_scraper"
	}
	if err := cassandra.InitCassandra(logger, cassandraHosts, keyspace); err != nil {
		logger.Fatalf("Failed to initialize Cassandra: %v", err)
	}
}
