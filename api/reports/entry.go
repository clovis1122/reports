package reports

import (
	"sort"
	"strconv"
	"time"
)

var tz = time.FixedZone("Dominican Republic Time", -4*60*60)

func convertToDominicanTimezone(date string) time.Time {
	parsedDate, _ := time.Parse(time.RFC3339, date)
	return parsedDate.In(tz)
}

func getDominicanStartOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tz)
}

func getDominicanEndOfDay(date time.Time) time.Time {
	return getDominicanStartOfDay(date.AddDate(0, 0, 1)).Add(time.Second * -1)
}

func getDominican9PMDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, tz)
}

func getDominican6AMDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, tz)
}

func getDominican12PMDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, tz)
}

// The RawEntry struct represents the data of a given entry.
type RawEntry struct {
	Pid         int64    `json:"pid"`
	Start       string   `json:"start"`
	Description string   `json:"description"`
	Stop        string   `json:"stop"`
	Tags        []string `json:"tags"`
}

func (e *RawEntry) toEntry() entry {
	return entry{
		ProjectID:   e.Pid,
		Tags:        e.Tags,
		Description: e.Description,
		Start:       convertToDominicanTimezone(e.Start),
		Stop:        convertToDominicanTimezone(e.Stop),
	}
}

// The underlying toggl entry.
type entry struct {
	ProjectID   int64
	Tags        []string
	Description string
	Start       time.Time
	Stop        time.Time
}

func (e *entry) trySplitInDifferentDays() []entry {
	if e.Start.Day() == e.Stop.Day() {
		return []entry{}
	}

	// Shallow Copy
	entry1, entry2 := *e, *e
	entry1.Stop = getDominicanEndOfDay(e.Start)
	entry2.Start = getDominicanStartOfDay(e.Start.AddDate(0, 0, 1))

	return []entry{entry1, entry2}
}

func (e *entry) trySplitAfter9PM() []entry {
	splitDate := getDominican9PMDate(e.Start)

	stopAfter9PM := e.Stop.Hour() >= splitDate.Hour()
	startBefore9PM := e.Start.Hour() < splitDate.Hour()

	if stopAfter9PM && startBefore9PM {
		// Shallow Copy
		entry1, entry2 := *e, *e
		entry1.Stop = splitDate.Add(time.Second * -1)
		entry2.Start = splitDate
		return []entry{entry1, entry2}
	}
	return []entry{}
}

func (e *entry) trySplitBefore6AM() []entry {
	splitDate := getDominican6AMDate(e.Start)

	startBefore6AM := e.Start.Hour() < splitDate.Hour()
	stopAfter6AM := e.Stop.Hour() >= splitDate.Hour()

	if startBefore6AM && stopAfter6AM {
		// Shallow Copy
		entry1, entry2 := *e, *e
		entry1.Stop = splitDate.Add(time.Second * -1)
		entry2.Start = splitDate

		return []entry{entry1, entry2}
	}
	return []entry{}
}

func (e *entry) trySplitBeforeSaturday12PM() []entry {
	splitDate := getDominican12PMDate(e.Start)

	if splitDate.Weekday() != 6 {
		return []entry{}
	}

	startBefore12PM := e.Start.Hour() < splitDate.Hour()
	stopAfter12PM := e.Stop.Hour() >= splitDate.Hour()

	if startBefore12PM && stopAfter12PM {
		// Shallow Copy
		entry1, entry2 := *e, *e
		entry1.Stop = splitDate.Add(time.Second * -1)
		entry2.Start = splitDate

		return []entry{entry1, entry2}
	}
	return []entry{}
}

func (e *entry) toAMFormat() string {
	return e.Start.Format(time.Kitchen) + " - " + e.Stop.Format(time.Kitchen)
}

func (e *entry) isNightly() bool {
	return e.Stop.Hour() <= getDominican6AMDate(e.Stop).Hour() || e.Stop.Hour() >= getDominican9PMDate(e.Start).Hour()
}

// Workday is Monday up to Saturday at 12PM.
func (e *entry) isWorkday() bool {
	if e.isWeekend() {
		return true
	}
	return e.Stop.Weekday() == 6 && e.Stop.Hour() < 12
}

func (e *entry) isWeekend() bool {
	return e.Start.Weekday() != 0 && e.Start.Weekday() != 6
}

func (e *entry) getDuration() float64 {
	return e.Stop.Sub(e.Start).Hours()
}

func (e *entry) formatDay() string {
	return e.Start.Format("Monday, 02/01/2006")
}

// EntryList contains several utilities to manage the entries.
type EntryList struct {
	Weekday        float64
	Weekend        float64
	NightlyWeekday float64
	NightlyWeekend float64
	entryMapping   map[int64]map[string][]entry
}

// AddEntry processes an item and adds it to the EntryList.
func (e *EntryList) AddEntry(item entry) {
	stack := []entry{item}

	for len(stack) > 0 {
		topEntry := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if entries := topEntry.trySplitInDifferentDays(); len(entries) > 0 {
			stack = append(stack, entries...)
			continue
		}

		if entries := topEntry.trySplitAfter9PM(); len(entries) > 0 {
			stack = append(stack, entries...)
			continue
		}

		if entries := topEntry.trySplitBefore6AM(); len(entries) > 0 {
			stack = append(stack, entries...)
			continue
		}

		if entries := topEntry.trySplitBeforeSaturday12PM(); len(entries) > 0 {
			stack = append(stack, entries...)
			continue
		}

		if e.entryMapping[topEntry.ProjectID] == nil {
			e.entryMapping[topEntry.ProjectID] = make(map[string][]entry)
		}

		e.entryMapping[topEntry.ProjectID][topEntry.formatDay()] = append(e.entryMapping[topEntry.ProjectID][topEntry.formatDay()], topEntry)

		if topEntry.isNightly() {
			if topEntry.isWorkday() {
				e.NightlyWeekday += topEntry.getDuration()
			} else {
				e.NightlyWeekend += topEntry.getDuration()
			}
		}

		if topEntry.isWorkday() {
			e.Weekday += topEntry.getDuration()
		} else {
			e.Weekend += topEntry.getDuration()
		}
	}
}

// GetProjectIDs returns the list of project IDs that are in the entry.
func (e *EntryList) GetProjectIDs() []int64 {
	var projectIDs []int64

	for projectID := range e.entryMapping {
		if projectID == 0 {
			continue
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs
}

// GetSummaryWithProjectInformation prints it.
func (e *EntryList) GetSummaryWithProjectInformation(projects map[int64]string) string {
	summary := "Summary: \n"
	summary += "\nTotal weekday: " + strconv.FormatFloat(e.Weekday, 'f', 2, 64)
	summary += "\nTotal weekend: " + strconv.FormatFloat(e.Weekend, 'f', 2, 64)
	summary += "\nTotal nightly (weekday): " + strconv.FormatFloat(e.NightlyWeekday, 'f', 2, 64)
	summary += "\nTotal nightly (weekend): " + strconv.FormatFloat(e.NightlyWeekend, 'f', 2, 64)

	for projectID, entryDaysMapping := range e.entryMapping {
		name := projects[projectID]

		summary += "\nProject name: " + name

		var datesKeys []string
		for key := range entryDaysMapping {
			datesKeys = append(datesKeys, key)
		}
		sort.Slice(datesKeys, func(i, j int) bool { return datesKeys[i] < datesKeys[j] })

		for _, day := range datesKeys {
			summary += "\n-" + day
			entries := entryDaysMapping[day]

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Start.Before(entries[j].Start)
			})
			for _, entry := range entries {
				var nightEntry string
				if entry.isNightly() {
					nightEntry = "[NIGHT]"
				} else {
					nightEntry = "[DAY]"
				}
				summary += "\n----" + entry.toAMFormat() + ": " + entry.Description + nightEntry
			}
		}

		summary += "\n"
	}

	return summary
}

// CreateFromRawEntries creates an entry list given some recently-fetched entries.
func CreateFromRawEntries(rawEntries []RawEntry) EntryList {
	entries := EntryList{
		entryMapping: make(map[int64]map[string][]entry),
	}
	for _, entry := range rawEntries {
		if entry.Tags != nil {
			for _, tag := range entry.Tags {
				if tag == "Extra" {
					entries.AddEntry(entry.toEntry())
				}
			}
		}
	}

	return entries
}
