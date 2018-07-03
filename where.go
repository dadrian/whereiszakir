package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/calendar/v3"
)

type zcache struct {
	Location        string
	When            time.Time
	DefaultLocation string
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *jwt.Config) *http.Client {
	c := config.Client(ctx)
	return c
}

func GetTimeFromEventTime(e *calendar.EventDateTime) (time.Time, error) {
	if e.DateTime != "" {
		return time.Parse(time.RFC3339, e.DateTime)
	} else {
		return time.Parse("2006-01-02", e.Date)
	}
}

func main() {

	var cid string
	var keyPath string
	var defaultLocation string
	flag.StringVar(&cid, "calendar", "primary", "Zakir travel calendar ID")
	flag.StringVar(&keyPath, "key", "where-is-zakir-key.json", "path to JSON key")
	flag.StringVar(&defaultLocation, "default-location", "Unknown", "displays when nothing is listed on the calendar")
	flag.Parse()

	ctx := context.Background()

	b, err := ioutil.ReadFile("where-is-zakir-key.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("unable to load jwt config: %v", err)
	}
	config.Subject = "davidcadrian@gmail.com"
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	cache := new(zcache)
	cache.DefaultLocation = defaultLocation

	handler := func(w http.ResponseWriter, r *http.Request) {
		obj := struct {
			Location string `json:"location"`
		}{
			Location: where(srv, cid, cache),
		}
		b, err := json.Marshal(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/where", handler)
	http.ListenAndServe("localhost:9720", nil)
}

func where(srv *calendar.Service, cid string, cache *zcache) string {

	n := time.Now()
	expiry := cache.When.Add(time.Minute)

	if n.Before(expiry) {
		return cache.Location
	}

	t := n.Format(time.RFC3339)
	events, err := srv.Events.List(cid).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()

	current := ""

	if err == nil && len(events.Items) > 0 {
		for _, i := range events.Items {
			start, err := GetTimeFromEventTime(i.Start)
			if err != nil {
				log.Fatalf("could not get start time: %v", err)
			}
			end, err := GetTimeFromEventTime(i.End)
			if err != nil {
				log.Fatalf("could not get end time: %v", err)
			}
			parts := strings.Split(i.Summary, "(")
			name := parts[0]
			name = strings.Trim(name, " ")
			if start.Before(n) && end.After(n) {
				current = name
			}
		}
	}

	if current == "" {
		current = cache.DefaultLocation
	}
	cache.Location = current
	cache.When = n
	return current
}
