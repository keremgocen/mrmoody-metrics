package storage

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type User struct {
	Mood string `json:"mood"`
	Date int64  `json:"date"`
}

type Campaign map[string]User
type Votes map[string]Campaign

// Setup creates a new InfluxDB client.
func Setup(user, pass, addr string) client.Client {

	// NOTE: this assumes you've setup a user and have setup shell env variables,
	// namely INFLUX_USER/INFLUX_PWD. If not just omit Username/Password below.
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: addr,
		// Username: os.Getenv("INFLUX_USER"),
		// Password: os.Getenv("INFLUX_PWD"),
	})
	if err != nil {
		log.Println("error creating client", err.Error())
	}

	return c
}

// CreateDB creates a new InfluxDB database with the given name.
func CreateDB(c client.Client, dbName string) error {

	s := []string{"CREATE DATABASE "}
	q := client.NewQuery(strings.Join(s, dbName), "", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		log.Println(response.Results)
	} else {
		log.Println("error creating database", dbName)
		return err
	}
	return nil
}

// WriteBatchPoints creates points using votes data and writes them into InfluxDB as a batch.
func WriteBatchPoints(c client.Client, p Votes, dbName string) error {
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbName,
		Precision: "s",
	})

	for k, value := range p {
		log.Println("Campaign:", k, "details:", value)
		tags := map[string]string{"campaign": k}

		for key2, user := range value {
			// Create a point and add to batch
			fields := map[string]interface{}{
				"user": key2,
				"mood": user.Mood,
			}

			pt, err := client.NewPoint("campaign", tags, fields, msToTime(strconv.FormatInt(user.Date, 10)))
			if err != nil {
				log.Println("error creating point", err.Error())
				return err
			}
			bp.AddPoint(pt)
		}
	}

	// Write the batch
	if err := c.Write(bp); err != nil {
		return err
	}

	return nil
}

// Query makes a static query.
func Query(c client.Client, dbName string) error {
	q := client.NewQuery("SELECT * FROM campaign", dbName, "s")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		log.Println("query results:", response.Results)
		return nil
	} else {
		return err
	}
}

func msToTime(ms string) time.Time {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		log.Println("error parsing duration", ms, err.Error())
		return time.Time{}
	}

	return time.Unix(0, msInt*int64(time.Millisecond))
}
