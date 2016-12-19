package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/keremgocen/mrmoody-metrics/storage"

	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	// read config file
	targetURL := viper.GetString("target_url")
	apiKey := viper.GetString("api_key")
	fmt.Printf("\n%s\n\n", viper.AllSettings())

	// setup storage
	c := storage.Setup()
	defer c.Close()
	storage.CreateDB()

	// create request to parse Firebase data
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authentication", apiKey)

	now := time.Now()
	nowStr := strconv.FormatInt(unixMilli(now), 10)
	fmt.Println("now:", now, " ts:", nowStr)
	lastMonth := now.AddDate(0, -3, 0)
	lastMonthStr := "1479495795942" //strconv.FormatInt(unixMilli(lastMonth), 10)
	fmt.Println("last month:", lastMonth, " ts:", lastMonthStr)

	q := req.URL.Query()
	// q.Add("orderBy", "\"date\"")
	// q.Add("startAt", lastMonthStr)
	// q.Add("endAt", nowStr)
	q.Add("print", "pretty")
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	// log outgoing request
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%q\n", dump)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error reading results from ", targetURL)
		fmt.Println("status ", resp.Status)
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// b = nil

	// fmt.Printf("%s\n", b)

	res := storage.Votes{}
	if err := json.Unmarshal([]byte(b), &res); err != nil {
		log.Fatal(err)
	}

	fmt.Println("res:", res)

	storage.WriteBatchPoints(c, res)
	storage.Query(c)

}

func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
