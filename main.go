package main

// https://github.com/google/google-api-go-client
// https://godoc.org/google.golang.org/api/calendar/v3#EventsService

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}

	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}

	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")

	os.MkdirAll(tokenCacheDir, 0700)

	return filepath.Join(tokenCacheDir, url.QueryEscape("calendar-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)

	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)
	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	list, err := srv.CalendarList.List().ShowHidden(false).Do()
	if err != nil || list == nil {
		log.Fatalf("Unable to retrieve user's calendars list. %v", err)
	}

	now := time.Now()

	monthBegin := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	fmt.Printf("%v - %v\n", monthBegin, monthEnd)

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	weekBegin := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	weekEnd := time.Date(now.Year(), now.Month(), now.Day()-weekday+7, 23, 59, 59, 0, now.Location())
	fmt.Printf("%v - %v\n", weekBegin, weekEnd)

	var lastStartTime, lastEndTime time.Time

	today := 0

	weekTotal := 0.0
	workDaysInWeek := 0
	monthTotal := 0.0
	workDaysInMonth := 0

	for _, cal := range list.Items {
		// fmt.Printf("%s%#v\n\n", strings.Repeat("=", 100), cal)

		// events, err := srv.Events.List(cal.Id).ShowDeleted(false).SingleEvents(true).TimeMin(monthBegin.Format(time.RFC3339)).TimeMax(monthEnd.Format(time.RFC3339)).OrderBy("startTime").Do()
		events, err := srv.Events.List(cal.Id).ShowDeleted(false).SingleEvents(true).TimeMin(time.Date(now.Year(), 0, 0, 0, 0, 0, 0, now.Location()).Format(time.RFC3339)).TimeMax(monthEnd.Format(time.RFC3339)).OrderBy("startTime").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
		}

		for _, ev := range events.Items {
			if !strings.HasPrefix(ev.Summary, "SolarWinds") {
				continue
			}

			startTime, err := time.Parse(time.RFC3339, ev.Start.DateTime)
			if err != nil {
				fmt.Println(err)
			}

			endTime, err := time.Parse(time.RFC3339, ev.End.DateTime)
			if err != nil {
				fmt.Println(err)
			}

			inWeek := false
			inMonth := false

			if startTime.Unix() > weekBegin.Unix() && endTime.Unix() < weekEnd.Unix() {
				inWeek = true
			}

			if startTime.Unix() > monthBegin.Unix() && endTime.Unix() < monthEnd.Unix() {
				inMonth = true
			}

			duration := endTime.Sub(startTime).Hours()

			if startTime.Day() != today {
				today = startTime.Day()

				lastStartTime = startTime
				lastEndTime = endTime

				if inWeek {
					workDaysInWeek += 1
				}

				if inMonth {
					workDaysInMonth += 1
				}
				// } else if startTime.Unix() < lastEndTime.Unix() {
				// 	duration = endTime.Sub(lastEndTime).Hours()
			}

			if inWeek {
				weekTotal += duration
			}

			if inMonth {
				monthTotal += duration
			}

			fmt.Printf("%v\t\t%v\t%v\n\n", startTime.Format(time.RFC1123), duration, ev.Summary)
		}
	}

	_ = lastStartTime
	_ = lastEndTime

	weekTargetTotal := float64(workDaysInWeek * 8)
	monthTargetTotal := float64(workDaysInMonth * 8)

	fmt.Printf("week total: %v of %v (%+.2f)\n", weekTotal, weekTargetTotal, -(weekTargetTotal - weekTotal))
	fmt.Printf("month total: %v of %v (%+.2f)\n", monthTotal, monthTargetTotal, -(monthTargetTotal - monthTotal))
}
