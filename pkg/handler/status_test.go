package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	// "strings"
	"testing"
	// "time"

	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/logger"
)

func TestStatus(t *testing.T) {

	var (
		assets      http.FileSystem = http.Dir("../resources")
		emptyAssets http.FileSystem = http.Dir("../empty")
	)

	flags := map[string]string{}
	logger := logger.SetupLogger("DEBUG")

	mux := http.NewServeMux()

	mux.HandleFunc("/", Status(assets, flags, "externalPathPrefix", logger).ServeHTTP)
	mux.HandleFunc("/empty", Status(emptyAssets, flags, "externalPathPrefix", logger).ServeHTTP)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	rootRoute, _ := url.Parse(ts.URL + "/")
	emptyRoute, _ := url.Parse(ts.URL + "/empty")

	// ---------------------------------------------------------------------------
	//  CASE: Getting route "/" without error
	// ---------------------------------------------------------------------------
	resp, _ := http.Get(rootRoute.String())
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status() : Test 1 FAILED, expected status code: 200, got: %d", resp.StatusCode)
	} else {
		t.Log("Status() : Test 1 PASSED.")
	}

	// ---------------------------------------------------------------------------
	//  CASE: Getting route "/empty"
	// ---------------------------------------------------------------------------
	resp, _ = http.Get(emptyRoute.String())
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Status() : Test 2 FAILED, expected status code: 500, got: %d", resp.StatusCode)
	} else {
		t.Log("Status() : Test 2 PASSED.")
	}

}

// func TestCount(t *testing.T) {
//
// 	data := &data{
// 		Birth:      time.Now(),
// 		BuildInfo:  map[string]string{"": ""},
// 		Flags:      map[string]string{"": ""},
// 		PathPrefix: "",
// 		counter:    int(10),
// 	}
//
// 	// ---------------------------------------------------------------------------
// 	//  CASE: Just count
// 	// ---------------------------------------------------------------------------
// 	count := data.Count()
// 	if count == int(11) {
// 		t.Logf("Count() : Test 1 PASSED, expect count='11' and got count='%v'", count)
// 	} else {
// 		t.Errorf("Count() : Test 1 FAILED, count='%v', expecting count='11'", count)
// 	}
//
// }

// func TestFormatTimestamp(t *testing.T) {
//
// 	data := &data{
// 		Birth:      time.Now(),
// 		BuildInfo:  map[string]string{"": ""},
// 		Flags:      map[string]string{"": ""},
// 		PathPrefix: "",
// 		counter:    int(10),
// 	}
// 	ts := int64(1000)
// 	tsn := time.Date(1970, 01, 01, 00, 00, 01, 00, time.UTC)
//
// 	// ---------------------------------------------------------------------------
// 	//  CASE: Testing timestamp
// 	// ---------------------------------------------------------------------------
// 	// fts := strings.Split(data.FormatTimestamp(ts), " +")[0]
// 	fts := data.FormatTimestamp(ts)
// 	if fts == tsn.String() {
// 		t.Logf("FormatTimestamp() : Test 1 PASSED, expect ts='1970-01-01 03:00:01' and got ts='%s'", fts)
// 	} else {
// 		t.Errorf("FormatTimestamp() : Test 1 FAILED, got ts='%s', tsn='%s'", fts, tsn)
// 	}
//
// }
