package alertapi

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const validPost = `{ 
	"receiver":"thereceiver",
	"status":"firing",
	"alerts":[ 
	   { 
		  "status":"firing",
		  "labels":{ 
			 "alertname":"TestAlert",
			 "instance":"localhost:9100",
			 "job":"node",
			 "monitor":"thereceiver",
			 "severity":"HIGH"
		  },
		  "annotations":{ 
			 "summary":"This is a test alert summary"
		  },
		  "startsAt":"2019-12-07T19:54:07.498629312+02:00",
		  "endsAt":"0001-01-01T00:00:00Z",
		  "generatorURL":"http://example.com:9090/graph?g0.expr=testURL",
		  "fingerprint":"c0075d874fc21056"
	   }
	],
	"groupLabels":{ 
 
	},
	"commonLabels":{ 
	   "alertname":"TestAlert",
	   "instance":"localhost:9100",
	   "job":"node",
	   "monitor":"thereceiver",
	   "severity":"HIGH"
	},
	"commonAnnotations":{ 
	   "summary":"This is a test alert summary"
	},
	"externalURL":"http://example.com:9093",
	"version":"4",
	"groupKey":"{}:{}"
 }
 `

func TestHealthz(t *testing.T) {
	l, _ := setupService(t)
	defer l.Close()
	resp, err := http.Get("http://" + l.Addr().String() + healthzPath)
	assert.NoError(t, err, "Could not reach healthz")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	testResponse(t, resp, okBody)
}

func TestAlertManagerGet(t *testing.T) {
	l, _ := setupService(t)
	defer l.Close()
	resp, err := http.Get("http://" + l.Addr().String() + alertmanagerPath)
	assert.NoError(t, err, "Could not reach alert manager")
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestAlertManagerEmptyPost(t *testing.T) {
	l, _ := setupService(t)
	defer l.Close()
	resp, err := http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", nil)
	assert.NoError(t, err, "Could not reach alert manager")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	testResponse(t, resp, "{\"status\":400,\"message\":\"unexpected end of JSON input\"}")
}

func TestAlertManagerValidJSON(t *testing.T) {
	l, api := setupService(t)
	defer l.Close()
	payload := bytes.NewBuffer([]byte(validPost))
	resp, err := http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", payload)
	readEntry := <-api.Alerts()
	assert.NoError(t, err, "Could not reach alert manager")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	testResponse(t, resp, "{\"status\":200,\"message\":\"success\"}")

	assert.Contains(t, "thereceiver", readEntry.Receiver)
	assert.Contains(t, "firing", readEntry.Status)
	assert.Equal(t, 1, len(readEntry.Alerts))
}

func TestAlertManagerTimeout(t *testing.T) {
	l, _ := setupService(t)
	defer l.Close()
	payload := bytes.NewBuffer([]byte(validPost))
	resp, err := http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", payload)
	assert.NoError(t, err, "Could not reach alert manager")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	payload = bytes.NewBuffer([]byte(validPost))
	resp, err = http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", payload)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
	testResponse(t, resp, "{\"status\":408,\"message\":\"Request Timeout\"}")

}

func TestAlertManagerRateLimit(t *testing.T) {
	l, api := setupService(t)
	defer l.Close()
	payload := bytes.NewBuffer([]byte(validPost))
	resp, err := http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", payload)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NoError(t, err, "Could not reach alert manager")
	_ = <-api.Alerts()

	payload = bytes.NewBuffer([]byte(validPost))
	resp, err = http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", payload)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = <-api.Alerts()

	payload = bytes.NewBuffer([]byte(validPost))
	resp, err = http.Post("http://"+l.Addr().String()+alertmanagerPath, "application/json", payload)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	testResponse(t, resp, "{\"status\":429,\"message\":\"Too Many Requests\"}")

}

func setupService(t *testing.T) (net.Listener, *AlertAPI) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	assert.NoError(t, err, "Could not start test server")
	api := New(1, 2, 1)
	go http.Serve(l, api)
	return l, api
}

func testResponse(t *testing.T, resp *http.Response, body string) {
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err, "Could not read response")
	assert.Equal(t, body, string(data), "Response not expected")
}
