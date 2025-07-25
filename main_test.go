package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain starts the server and runs all the tests.
// By doing this, you can run **actual** integration tests without starting the server.
func TestMain(m *testing.M) {
	flag.Parse() // NOTE: this is needed to parse args from go test command

	// port := func() string { // Get a free port to run the server
	// 	listener, err := net.Listen("tcp", ":0")
	// 	if err != nil {
	// 		log.Fatalf("failed to listen: %v", err)
	// 	}
	// 	defer listener.Close()
	// 	addr := listener.Addr().(*net.TCPAddr)
	// 	return strconv.Itoa(addr.Port)
	// }()

	port := "9200" //Hard Coded Port

	go func() { // Start the server in a goroutine
		if err := run(os.Stdout, []string{"test", "--port", port}); err != nil {
			log.Fatal(err)
		}
	}()

	endpoint = "http://localhost:" + port

	start := time.Now() // wait for server to be healthy before tests.
	for time.Since(start) < 3*time.Second {
		if res, err := http.Get(endpoint + "/health"); err == nil && res.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

// endpoint holds the server endpoint started by TestMain, not intended to be updated.
var endpoint string

// TestGetHealth tests the /health endpoint.
// Server is started by [TestMain] so that the test can make requests to it.
func TestGetHealth(t *testing.T) {
	t.Parallel()
	// response is repeated, but this describes intention of test better.
	// For example, you can add fields only needed for testing.
	type response struct {
		Version  string    `json:"version"`
		Revision string    `json:"vcs.revision"`
		Time     time.Time `json:"vcs.time"`
		// Modified bool      `json:"vcs.modified"`
	}

	// actual http request to the server.
	res, err := http.Get(endpoint + "/health")
	testNil(t, err)
	t.Cleanup(func() {
		err = res.Body.Close()
		testNil(t, err)
	})
	testEqual(t, http.StatusOK, res.StatusCode)
	testEqual(t, "application/json", res.Header.Get("Content-Type"))
	testNil(t, json.NewDecoder(res.Body).Decode(&response{}))
}

func testEqual[T comparable](tb testing.TB, want, got T) {
	tb.Helper()
	if want != got {
		tb.Fatalf("want: %v; got: %v", want, got)
	}
}

func testNil(tb testing.TB, err error) {
	tb.Helper()
	testEqual(tb, nil, err)
}

func testContains(tb testing.TB, needle string, haystack string) {
	tb.Helper()
	if !strings.Contains(haystack, needle) {
		tb.Fatalf("%q not in %q", needle, haystack)
	}
}

//Following Tests need to be updated if needed/wanted
//
//
//
//
// TestHelloWorld tests the /helloworld endpoint.
// You can add more test as needed without starting the server again.
// func TestGetHelloWorld(t *testing.T) {
// 	t.Parallel()
// 	res, err := http.Get(endpoint + "/helloworld")
// 	testNil(t, err)
// 	testEqual(t, http.StatusOK, res.StatusCode)
// 	testEqual(t, "application/json", res.Header.Get("Content-Type"))

// 	sb := strings.Builder{}
// 	_, err = io.Copy(&sb, res.Body)
// 	testNil(t, err)
// 	t.Cleanup(func() {
// 		err = res.Body.Close()
// 		testNil(t, err)
// 	})

// 	testContains(t, "Hello World", sb.String())
// 	testContains(t, "Uptime", sb.String())
// }

// TestAccessLogMiddleware tests accesslog middleware
// func TestAccessLogMiddleware(t *testing.T) {
// 	t.Parallel()

// 	type record struct {
// 		Method string `json:"method"`
// 		Path   string `json:"path"`
// 		Query  string `json:"query"`
// 		Status int    `json:"status_code"`
// 		body   []byte `json:"-"`
// 		Bytes  int    `json:"size_bytes"`
// 	}

// 	tests := []record{
// 		{
// 			Method: "GET",
// 			Path:   "/test",
// 			Query:  "?key=value",
// 			Status: http.StatusOK,
// 			body:   []byte(`{"hello":"world"}`),
// 		},
// 		{
// 			Method: "POST",
// 			Path:   "/api",
// 			Status: http.StatusCreated,
// 			body:   []byte(`{"id":1}`),
// 		},
// 		{
// 			Method: "DELETE",
// 			Path:   "/users/1",
// 			Status: http.StatusNoContent,
// 		},
// 	}

// 	for _, tt := range tests {
// 		name := strings.Join([]string{tt.Method, tt.Path, tt.Query, strconv.Itoa(tt.Status)}, " ")
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()

// 			var buffer strings.Builder
// 			handler := middleware.AccessLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
// 				w.WriteHeader(tt.Status)
// 				w.Write(tt.body) //nolint:errcheck
// 			}), zerolog.New(&buffer))

// 			req := httptest.NewRequest(tt.Method, tt.Path+tt.Query, bytes.NewReader(tt.body))
// 			rec := httptest.NewRecorder()
// 			handler.ServeHTTP(rec, req)

// 			fmt.Println(buffer.String())

// 			var log record
// 			err := json.NewDecoder(strings.NewReader(buffer.String())).Decode(&log)
// 			testNil(t, err)

// 			fmt.Println(log)

// 			testEqual(t, tt.Method, log.Method)
// 			testEqual(t, tt.Path, log.Path)
// 			testEqual(t, strings.TrimPrefix(tt.Query, "?"), log.Query)
// 			testEqual(t, len(tt.body), log.Bytes)
// 			testEqual(t, tt.Status, log.Status)
// 		})
// 	}
// }
