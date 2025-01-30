package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"mono-golang/config"
	"mono-golang/internal/domain/book"
	calendarPb "mono-golang/internal/domain/calendar/proto"
)

// Version is injected using ldflags during build time
const Version = "v0.1.0"

var url = ""

func main() {
	log.Printf("Starting e2e API version: %s\n", Version)
	cfg := config.New()

	url = fmt.Sprintf("http://%s:%s", cfg.Api.Host, cfg.Api.Port)

	waitForApi(fmt.Sprintf("%s/api/health/readiness", url))

	run()
}

func run() {
	testBook()
	testCalendar()

	log.Println("all tests have passed.")
}

func testCalendar() {
	testCreateEvent()
	testListEvents()
	testUpdateEvent()
	testDeleteEvent()
}

func testCreateEvent() {
	event := &calendarPb.CreateEventRequest{
		Event: &calendarPb.Event{
			Title:       "Test Event",
			Description: "This is a test event",
			Location:    "Test Location",
			StartTime:   time.Now().Unix(),
			EndTime:     time.Now().Add(2 * time.Hour).Unix(),
			Attendees:   []string{"test@example.com"},
			Status:      "pending",
		},
		UserId: "test-user",
	}

	bR, _ := json.Marshal(event)

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/calendar/events", url),
		"application/json",
		bytes.NewBuffer(bR),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusCreated, resp.StatusCode)
	}

	var response calendarPb.CreateEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalln(err)
	}

	if response.Event == nil {
		log.Fatalln("expected event in response, got nil")
	}

	log.Println("testCreateEvent passes")
}

func testListEvents() {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/calendar/events", url))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusOK, resp.StatusCode)
	}

	var response calendarPb.ListEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalln(err)
	}

	log.Println("testListEvents passes")
}

func testUpdateEvent() {
	event := &calendarPb.UpdateEventRequest{
		Event: &calendarPb.Event{
			Id:          "test-event-id",
			Title:       "Updated Test Event",
			Description: "This is an updated test event",
			Location:    "Updated Test Location",
			StartTime:   time.Now().Unix(),
			EndTime:     time.Now().Add(3 * time.Hour).Unix(),
			Status:      "confirmed",
		},
		UserId: "test-user",
	}

	bR, _ := json.Marshal(event)

	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/calendar/events/%s", url, event.Event.Id),
		bytes.NewBuffer(bR),
	)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusOK, resp.StatusCode)
	}

	var response calendarPb.UpdateEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalln(err)
	}

	log.Println("testUpdateEvent passes")
}

func testDeleteEvent() {
	eventID := "test-event-id"
	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/calendar/events/%s", url, eventID),
		nil,
	)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusOK, resp.StatusCode)
	}

	log.Println("testDeleteEvent passes")
}

func testBook() {
	testEmptyBook()
	id := testAddOneBook()
	id = testGetOneBook(id)
	testUpdateBook(id)
	testDeleteOneBook(id)
}

func testEmptyBook() {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/book", url))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if status := resp.StatusCode; status != http.StatusOK {
		log.Printf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected, _ := json.Marshal(make([]*book.Res, 0))

	if !bytes.Equal(expected, got) {
		log.Printf("handler returned unexpected body: got %v want %v", string(got), expected)
	}

	log.Println("testEmptyBook passes")
}

func testAddOneBook() uint64 {
	want := &book.CreateRequest{
		Title:         "test01",
		PublishedDate: "2020-02-02",
		ImageURL:      "https://example.com/image.png",
		Description:   "test01",
	}

	bR, _ := json.Marshal(want)

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/book", url),
		"Content-Type: application/json",
		bytes.NewBuffer(bR),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	gotBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	got := book.Res{}
	err = json.Unmarshal(gotBody, &got)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Printf("error code want %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if want.Title != got.Title && want.Description != got.Description && want.
		ImageURL != got.ImageURL && want.PublishedDate != got.PublishedDate.String() {
		log.Printf("want %v, got %v\n", want, got)
	}

	log.Println("testAddOneBook passes")
	return got.ID
}

func testGetOneBook(id uint64) uint64 {
	client := &http.Client{}

	url := fmt.Sprintf("%s/api/v1/book/%d", url, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusOK, resp.StatusCode)
	}

	got := book.Res{}
	err = json.Unmarshal(respBody, &got)
	if err != nil {
		log.Println(err)
	}

	log.Println("testGetBook passes")

	return got.ID
}

func testUpdateBook(bookID uint64) {
	newBook := book.CreateRequest{
		Title:         "updated title",
		PublishedDate: "2020-07-31T15:04:05.123499999Z",
		ImageURL:      "https://example.com/image.png",
		Description:   "test description",
	}

	client := &http.Client{}

	bR, err := json.Marshal(&newBook)
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("%s/api/v1/book/%d", url, bookID)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bR))
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusOK, resp.StatusCode)
	}

	got := book.Res{}
	err = json.Unmarshal(respBody, &got)
	if err != nil {
		log.Println(err)
	}

	if got.ID != bookID && got.Title != newBook.Title && got.Description != newBook.Description && got.ImageURL != newBook.ImageURL {
		if err != nil {
			log.Fatalf("returned resource does not match. want %v, got %v", respBody, got)
		}
	}

	log.Println("testUpdateBook passes")
}

func testDeleteOneBook(id uint64) {
	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%s/api/v1/book/%d", url, id),
		nil,
	)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error code fail, want %d, got %d\n", http.StatusOK, resp.StatusCode)
	}
	log.Println("testDeleteOneBook passes")
}

func waitForApi(readinessURL string) {
	log.Println("Connecting to api with exponential backoff... ")
	for {
		//nolint:gosec
		_, err := http.Get(readinessURL)
		if err == nil {
			log.Println("api is up")
			return
		}

		base, capacity := time.Second, time.Minute
		for backoff := base; err != nil; backoff <<= 1 {
			if backoff > capacity {
				backoff = capacity
			}

			// A pseudo-random number generator here is fine. No need to be
			// cryptographically secure. Ignore with the following comment:
			/* #nosec */
			jitter := rand.Int63n(int64(backoff * 3))
			sleep := base + time.Duration(jitter)
			time.Sleep(sleep)
			//nolint:gosec
			_, err := http.Get(readinessURL)
			if err == nil {
				log.Println("api is up")
				return
			}
		}
	}
}
