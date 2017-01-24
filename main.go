package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keremgocen/mrmoody-metrics/storage"

	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("No configuration file loaded - using defaults")
		// TODO defaults
	}

	// read config file
	targetURL := viper.GetString("target_url")
	apiKey := viper.GetString("api_key")
	period := viper.GetInt("update_period_minutes")
	dbAddr := viper.GetString("db_address")
	dbName := viper.GetString("db_name")
	dbUser := viper.GetString("db_user")
	dbPass := viper.GetString("db_pass")
	log.Printf("\n%s\n\n", viper.AllSettings())

	log.Println("Fetching period is set to", period, "seconds.")

	// setup storage
	c := storage.Setup(dbUser, dbPass, dbAddr)
	defer c.Close()
	err = storage.CreateDB(c, dbName)
	if err != nil {
		log.Println("failed to create database", err.Error())
	}

	// DEBUG TODO use seconds
	dur := time.Second * time.Duration(period)
	ticker := time.NewTicker(dur)
	done := make(chan bool, 1)
	go func(done chan bool) {
		for t := range ticker.C {
			log.Println("tick", period, "at", t)

			b, err := fetchFirebaseData(targetURL, apiKey)
			if err != nil {
				log.Fatal(err)
				break
			}

			res := storage.Votes{}
			if err := json.Unmarshal([]byte(b), &res); err != nil {
				log.Fatal(err)
				break
			}

			log.Println("res:", res)

			err2 := storage.WriteBatchPoints(c, res, dbName)
			if err2 != nil {
				log.Println("error writing points in db", err2.Error())
			}
			done <- true
		}
	}(done)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-stop)
	ticker.Stop()
	log.Println("Ticker stopped! Waiting for fetching routine..")

	// Block until we receive a notification from the worker on the channel.
	<-done

	err = storage.Query(c, dbName)
	if err != nil {
		log.Println("error querying db", err.Error())
	}

}

func fetchFirebaseData(URL, key string) ([]byte, error) {

	// create request to parse Firebase data
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authentication", key)

	// now := time.Now()
	// nowStr := strconv.FormatInt(unixMilli(now), 10)
	// log.Println("now:", now, " ts:", nowStr)
	// lastMonth := now.AddDate(0, -3, 0)
	// lastMonthStr := "1479495795942" //strconv.FormatInt(unixMilli(lastMonth), 10)
	// log.Println("last month:", lastMonth, " ts:", lastMonthStr)

	q := req.URL.Query()
	// q.Add("orderBy", "\"date\"")
	// q.Add("startAt", lastMonthStr)
	// q.Add("endAt", nowStr)
	// q.Add("print", "pretty")
	req.URL.RawQuery = q.Encode()

	// log outgoing request
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Println("Outgoing request", string(dump))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("error reading results from", URL)
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
