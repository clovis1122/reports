package reports

import (
	"encoding/json"
	"handler/requests"
	"strconv"
	"time"
)

const (
	togglEntryURL   = "https://www.toggl.com/api/v8/time_entries"
	togglProjectURL = "https://www.toggl.com/api/v8/projects"
)

type projectResponse struct {
	Data struct {
		Name string
	}
}

// FetchEntries fetches a list of raw entries from toggl.
func FetchEntries(token string) ([]RawEntry, error) {
	timeAgo := time.Now().AddDate(0, 0, -45).Format(time.RFC3339)
	body, err := requests.Get(togglEntryURL+"?start_date="+timeAgo, requests.SetBasicAuth(token, "api_token"))

	if err != nil {
		return []RawEntry{}, err
	}

	var rawEntries []RawEntry
	json.Unmarshal(body, &rawEntries)
	return rawEntries, nil
}

// FetchProjectIds fetches the names of the given projects.
func FetchProjectIds(token string, projectIds []int64) (map[int64]string, error) {
	projectMapping := map[int64]string{}
	var resp projectResponse

	for _, projectID := range projectIds {
		body, err := requests.Get(togglProjectURL+"/"+strconv.FormatInt(projectID, 10), requests.SetBasicAuth(token, "api_token"))
		if err != nil {
			return projectMapping, err
		}
		if err = json.Unmarshal(body, &resp); err != nil {
			return projectMapping, err
		}
		projectMapping[projectID] = resp.Data.Name
	}
	return projectMapping, nil
}
