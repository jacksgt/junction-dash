package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{} // use default options

var DB *sql.DB
var RSSITHRESHOLD = -50
var STATIONS = []string{"ant", "bear", "cheetah", "dolphin"}
var REQUESTS map[string]int
var USERS_COMPLETED int
var USERS_STARTED map[string]bool

func wifi_week(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade failed: %s\n", err)
		return
	}
	defer c.Close()
	var currentTimestamp int64
	for {
		// // receive message from client
		// mt, message, err := c.ReadMessage()
		// if err != nil {
		// 	log.Println("read:", err)
		// 	break
		// }
		// log.Printf("Got message %d %s\n", mt, message)

		// query database
		var lastseen time.Time
		var coord_x, coord_y float64
		rows, err := DB.Query("SELECT timestamp, lastseen,coordinate_x,coordinate_y FROM wifi_week WHERE confidence > 50 AND timestamp > $1 ORDER BY Timestamp LIMIT 1000", currentTimestamp)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&currentTimestamp, &lastseen, &coord_x, &coord_y)
			if err != nil {
				log.Fatal(err)
			}
			data := fmt.Sprintf("%s,%f,%f\n", lastseen, coord_x, coord_y)
			c.WriteMessage(websocket.TextMessage, []byte(data))
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getMacFromUrl(r *http.Request) string {
	fields := strings.Split(r.URL.String(), "/")
	macRaw := fields[len(fields)-1]
	mac, _ := url.QueryUnescape(macRaw)
	return strings.ToLower(mac)
}

func getRssi(station string, mac string) int {
	//	var timestamp time.Time
	var rssi int
	rows, err := DB.Query("SELECT MAX(rssi) AS rssi FROM wifi_diy WHERE station = $1 AND mac = $2 AND ts > NOW() - INTERVAL '1 minute' LIMIT 1", station, mac)
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			rows.Scan(&rssi)
		}
	}
	if rssi == 0 {
		rssi = -999
	}
	return rssi
}

func completed(station string, mac string) bool {
	var timestamp time.Time
	err := DB.QueryRow("SELECT ts FROM wifi_diy WHERE station = $1 AND mac = $2 AND rssi > $3 LIMIT 1", station, mac, RSSITHRESHOLD).Scan(&timestamp)
	if err != nil {
		//log.Printf("%s\n", err)
		return false
	}
	return true
}

func serve_mac(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %s\n", err)
		return
	}
	defer c.Close()

	//	c.SetWriteDeadline(time.Now().Add(30 * time.Second))

	mac := getMacFromUrl(r)
	USERS_STARTED[mac] = true
	for {
		values := make(map[string]int)
		for _, s := range STATIONS {
			var stationsCompleted int
			if completed(s, mac) {
				fmt.Printf("MAC %s completed %s\n", mac, s)
				msg := fmt.Sprintf(`{"type":"complete","station":"%s"}`, s)
				c.WriteMessage(websocket.TextMessage, []byte(msg))
				stationsCompleted += 1
			}
			values[s] = getRssi(s, mac)
			if stationsCompleted == len(STATIONS) {

			}
		}

		data, _ := json.Marshal(map[string]interface{}{
			"type":   "update",
			"values": values,
		})
		err := c.WriteMessage(websocket.TextMessage, []byte(data))
		if err != nil {
			// return
		}

		time.Sleep(1 * time.Second)
	}
}

func mock_sensors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(
		w,
		`{"ant":"Ant", "bear":"Bear","cheetah":"Cheetah","dolphin":"Dolphin"}`,
	)
}

func request_sound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodPost {
		station := r.FormValue("station")
		if station != "" {
			REQUESTS[station] += 1
			fmt.Fprintf(w, "%d\n", REQUESTS[station])
		}
		return
	}

	// serve get request
	fields := strings.Split(r.URL.String(), "/")
	stationRaw := fields[len(fields)-1]
	fmt.Fprintf(w, "%d\n", REQUESTS[strings.ToLower(stationRaw)])
}

func serve_stats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,
		`{"users_completed":%d,"users_started":%d}`,
		USERS_COMPLETED,
		len(USERS_STARTED),
	)
}

func main() {
	connStr := "postgres://avnadmin:yqt245bcfy6o5xmo@pg-6ef61e2-aalto-2dd2.aivencloud.com:28694/defaultdb?sslmode=require"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	REQUESTS = make(map[string]int)
	USERS_STARTED = make(map[string]bool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//	http.HandleFunc("/wifi_week", wifi_week)
	http.HandleFunc("/mac/", serve_mac)
	http.HandleFunc("/sensor/", mock_sensors)
	http.HandleFunc("/sound/", request_sound)
	http.HandleFunc("/stats/", serve_stats)
	http.Handle("/", http.FileServer(http.Dir(".")))
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
