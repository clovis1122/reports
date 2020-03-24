package reports

// GetTogglReport generates the report.
func GetTogglReport(token string) (string, error) {
	rawEntries, err := FetchEntries(token)
	if err != nil {
		return "", err
	}
	entryList := CreateFromRawEntries(rawEntries)
	projectIds := entryList.GetProjectIDs()
	projectMapping, err := FetchProjectIds(token, projectIds)
	if err != nil {
		return "", err
	}
	summary := entryList.GetSummaryWithProjectInformation(projectMapping)
	return summary, nil
}
