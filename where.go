package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

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

type ServiceAccountKey struct {
	ClientEmail string `json:"client_email,omitempty"`
}

func main() {

	var cid string
	var keyPath string
	var defaultLocation string
	var verbose bool
	flag.StringVar(&cid, "calendar", "primary", "Zakir travel calendar ID")
	flag.StringVar(&keyPath, "key", "where-is-zakir-key.json", "path to JSON key")
	flag.StringVar(&defaultLocation, "default-location", "Unknown", "displays when nothing is listed on the calendar")
	flag.BoolVar(&verbose, "verbose", false, "debug-level logging")
	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	ctx := context.Background()

	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %s", err)
	}
	var k ServiceAccountKey
	if err := json.Unmarshal(b, &k); err != nil {
		log.Fatalf("Could not deserialize key json: %s", err)
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
		log.Debugf("using cache (expiry %d)", expiry)
		return cache.Location
	}

	t := n.Format(time.RFC3339)
	events, err := srv.Events.List(cid).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()

	current := ""

	if err == nil && len(events.Items) > 0 {
		log.Debugf("got %d events", len(events.Items))
		for _, i := range events.Items {
			log.Debugf("checking event \"%s\"", i.Summary)
			start, err := GetTimeFromEventTime(i.Start)
			if err != nil {
				log.Fatalf("could not get start time: %v", err)
			}
			log.Debugf("start: %d", start)
			end, err := GetTimeFromEventTime(i.End)
			if err != nil {
				log.Fatalf("could not get end time: %v", err)
			}
			log.Debugf("end: %d", end)
			parts := strings.Split(i.Summary, "(")
			name := parts[0]
			name = strings.Trim(name, " ")
			if start.Before(n) && end.After(n) {
				log.Debug("using event as current!")
				current = name
			}
		}
	} else if err != nil {
		log.Errorf("error talking to calendar: %s", err)
	} else {
		log.Debug("no events received")
	}

	if current == "" {
		log.Debugf("couldn't find event, using default %s", cache.DefaultLocation)
		current = cache.DefaultLocation
	}
	cache.Location = current
	cache.When = n
	return current
}
