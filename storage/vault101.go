package storage

import (
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	MyDB     = "test"
	username = "kerem"
	password = "test"
)

type User struct {
	Mood string `json:"mood"`
	Date int64  `json:"date"`
}

type Campaign map[string]User
type Votes map[string]Campaign

// Create a new client
func Setup() client.Client {
	// NOTE: this assumes you've setup a user and have setup shell env variables,
	// namely INFLUX_USER/INFLUX_PWD. If not just omit Username/Password below.
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
		// Username: os.Getenv("INFLUX_USER"),
		// Password: os.Getenv("INFLUX_PWD"),
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}

	return c
}

// Create a Database with a query
func CreateDB() {
	// Make client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	q := client.NewQuery("CREATE DATABASE test", "", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	}
}

// Create a batch and add a point
func WriteBatchPoints(c client.Client, p Votes) {
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "test",
		Precision: "s",
	})

	for k, value := range p {
		fmt.Println("Campaign:", k, "details:", value)
		tags := map[string]string{"campaign": k}

		for key2, user := range value {
			fmt.Println("User:", key2, "date:", user.Date, "mood:", user.Mood)

			// Create a point and add to batch
			fields := map[string]interface{}{
				"user": key2,
				"mood": user.Mood,
			}

			pt, err := client.NewPoint("campaign", tags, fields, msToTime(strconv.FormatInt(user.Date, 10)))
			if err != nil {
				fmt.Println("Error: ", err.Error())
			}
			bp.AddPoint(pt)
		}
	}

	// Write the batch
	c.Write(bp)
}

// Make a Query
func Query(c client.Client) {
	q := client.NewQuery("SELECT * FROM campaign", "test", "s")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	}
}

func msToTime(ms string) time.Time {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		fmt.Println("Failed to parse", ms)
		return time.Time{}
	}

	// fmt.Println(time.Unix(0, msInt*int64(time.Millisecond)))

	return time.Unix(0, msInt*int64(time.Millisecond))
}
