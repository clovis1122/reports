package reports

import (
	"context"
	"crypto/tls"
	"handler/requests"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testEntryResponse = `[
		{
			"id": 123,
			"guid": "fakeguid",
			"wid": 123,
			"pid": 123,
			"tid": 123,
			"billable": true,
			"start": "2019-11-15T21:03:12+00:00",
			"stop": "2019-11-16T15:16:03+00:00",
			"duration": 65571,
			"description": "This is a fake task",
			"duronly": false,
			"at": "2019-11-16T15:16:03+00:00",
			"uid": 123,
			"tags": ["Extra"]
		}
	]`
	testProjectResponse = `
	{
		"data": {
			"id": 123,
			"wid": 123,
			"cid": 123,
			"name": "fakeproject",
			"billable": true,
			"is_private": true,
			"active": true,
			"template": false,
			"at": "2019-05-21T13:30:29+00:00",
			"created_at": "2018-02-12T18:09:17+00:00",
			"color": "14",
			"auto_estimates": true,
			"actual_hours": 123,
			"hex_color": "#000000"
		}
	}`
)

// Neat trick to test external requests, see: https://itnext.io/how-to-stub-requests-to-remote-hosts-with-go-6c2c1db32bf2
func getTestingClient(handler http.Handler) (http.Client, func()) {
	s := httptest.NewTLSServer(handler)
	cli := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	return cli, s.Close
}

func TestReportSummary(t *testing.T) {
	faketoken := "faketoken"
	responses := []string{testEntryResponse, testProjectResponse}
	cnt := 0

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pwd, _ := r.BasicAuth()

		if user != faketoken {
			t.Error("Basic Auth user does not match. Expected: " + faketoken + ", Got: " + user)
		}
		if pwd != "api_token" {
			t.Error("Basic Auth pwd does not match. Expected: api_token, Got: " + pwd)
		}

		w.Write([]byte(responses[cnt]))
		cnt++
	})
	client, close := getTestingClient(h)
	defer close()
	requests.SetClient(client)
	report, err := GetTogglReport(faketoken)
	t.Log(report)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(report, "Summary:") {
		t.Error("Does not contain summary")
	}
	if !strings.Contains(report, "Project name: fakeproject") {
		t.Error("Does not contain Project name: fakeproject")
	}
}
