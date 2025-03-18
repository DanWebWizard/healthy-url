package cassandra

import (
	"fmt"
	"time"

	"healthy-url/models"

	"github.com/gocql/gocql"
	"github.com/sirupsen/logrus"
)

var session *gocql.Session

// InitCassandra initializes the connection to the Cassandra database and creates the keyspace and table if they do not exist.
// Returns an error if the operation fails.
func InitCassandra(logger *logrus.Logger, hosts []string, keyspace string) error {
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second

	// Retry creating the temporary session up to 10 times
	var tempSession *gocql.Session
	var err error
	for i := 0; i < 10; i++ {
		tempCluster := *cluster
		tempCluster.Keyspace = "system"
		tempSession, err = tempCluster.CreateSession()
		if err == nil {
			logger.Println("Successfully connected to Cassandra system keyspace.")
			break
		}
		logger.Warnf("Failed to connect to Cassandra system keyspace, retrying in 15 seconds... (%d/10)", i+1)
		time.Sleep(15 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra system keyspace after 10 attempts: %v", err)
	}
	defer tempSession.Close()

	// Create keyspace if it doesn't exist
	err = tempSession.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {
		'class' : 'SimpleStrategy', 'replication_factor' : 1 };`, keyspace)).Exec()
	if err != nil {
		return fmt.Errorf("failed to create keyspace: %v", err)
	}

	// Update the cluster configuration to use the new keyspace
	cluster.Keyspace = keyspace

	// Retry creating the main session up to 10 times
	for i := 0; i < 10; i++ {
		session, err = cluster.CreateSession()
		if err == nil {
			logger.Println("Successfully connected to Cassandra.")
			break
		}
		logger.Warnf("Failed to connect to Cassandra, retrying in 15 seconds... (%d/10)", i+1)
		time.Sleep(15 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra after 10 attempts: %v", err)
	}

	// Create table if it doesn't exist
	err = session.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS url_healths (
		id UUID PRIMARY KEY,
		url text,
		healthy_url boolean,
		unhealthy_url boolean,
		unreachable_url boolean,
		time_taken int,
		status_code int
	);`)).Exec()
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// CloseCassandra closes the connection to the Cassandra database.
func CloseCassandra() {
	if session != nil {
		session.Close()
	}
}

// InsertUrlHealth inserts a UrlHealth record into the Cassandra database.
// Returns an error if the operation fails.
func InsertUrlHealth(logger *logrus.Logger, urlHealth models.UrlHealth) error {
	logger.Printf("Inserting url health: URL=%s, HealthyUrl=%t, UnhealthyUrl=%t, UnreachableUrl=%t, TimeTaken=%d, StatusCode=%d",
		urlHealth.URL, urlHealth.HealthyUrl, urlHealth.UnhealthyUrl, urlHealth.UnreachableUrl, urlHealth.TimeTaken, urlHealth.StatusCode)
	if err := session.Query(`INSERT INTO url_healths (id, url, healthy_url, unhealthy_url, unreachable_url, time_taken, status_code) VALUES (uuid(), ?, ?, ?, ?, ?, ?)`,
		urlHealth.URL, urlHealth.HealthyUrl, urlHealth.UnhealthyUrl, urlHealth.UnreachableUrl, urlHealth.TimeTaken, urlHealth.StatusCode).Exec(); err != nil {
		logger.Errorf("Failed to insert url_healths: %v", err)
		return err
	}
	logger.Println("Successfully inserted url health into Cassandra.")
	return nil
}
