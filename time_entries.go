package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

var (
	// filterTimeEntriesBefore is a filter to get time entries between two dates.
	filterTimeEntriesBefore = "[{\"user\":{\"operator\":\"=\",\"values\":[\"%d\"]}},{\"spent_on\":{\"operator\":\"<>d\",\"values\": [\"%s\",\"%s\"]}}]\n"

	// filterTimeEntriesWorkPackage is a filter to get all the time entries for a given work package.
	filterTimeEntriesWorkPackage = "[{\"work_package\":{\"operator\":\"=\",\"values\": [\"%d\"]}}]\n"
)

// TimeEntryCollection represents a collection of time entries.
type TimeEntryCollection struct {
	Embedded struct {
		Elements []TimeEntry `json:"elements"`
	} `json:"_embedded"`
}

// TimeEntry represents a single time log entry.
type TimeEntry struct {
	Id      int `json:"id"`
	Comment struct {
		Raw string `json:"raw"`
	} `json:"comment"`
	Hours string `json:"hours"`
	Date  string `json:"spentOn"`
	Links struct {
		WorkPackage struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		}
	} `json:"_links"`
}

// TimeEntryRequest represents a request to create a new time entry.
type TimeEntryRequest struct {
	Comment struct {
		Raw string `json:"raw"`
	} `json:"comment"`
	Hours    string `json:"hours"`
	Date     string `json:"spentOn"`
	Activity struct {
		Href  string `json:"href"`
		Title string `json:"title"`
	} `json:"activity"`
	Links struct {
		WorkPackage struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"workPackage"`
	} `json:"_links"`
	User struct {
		Href string `json:"href"`
	} `json:"user"`
}

// DeleteTimeEntry deletes a time entry.
func (c *Client) DeleteTimeEntry(id int) error {
	endpoint := fmt.Sprintf("%stime_entries/%d", c.baseURL, id)
	_, err := c.doRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	return nil
}

// CreateTimeEntry creates a new time entry for a given user.
func (c *Client) CreateTimeEntry(te *TimeEntryRequest) error {
	endpoint := fmt.Sprintf("%stime_entries", c.baseURL)
	jsonValue, err := json.Marshal(te)
	if err != nil {
		return fmt.Errorf("error marshalling request: %v", err)
	}
	_, err = c.doRequest("POST", endpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	return nil
}

// UpdateTimeEntryDuration updates the duration of a time entry.
func (c *Client) UpdateTimeEntryDuration(timeEntryId int, duration string, comment string, spendOn string) error {
	endpoint := fmt.Sprintf("%stime_entries/%d", c.baseURL, timeEntryId)
	update := map[string]interface{}{"hours": duration, "comment": map[string]string{"raw": comment}, "spentOn": spendOn}
	jsonValue, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("error marshalling request: %v", err)
	}
	_, err = c.doRequest("PATCH", endpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	return nil
}

// ListTimeEntries returns a collection of time entries for a given work package.
func (c *Client) ListTimeEntries(workPackageId int) (*TimeEntryCollection, error) {
	filters := fmt.Sprintf(filterTimeEntriesWorkPackage, workPackageId)
	return c.listTimeEntries(filters)
}

// ListTimeEntriesBefore returns a collection of time entries from the last n days.
func (c *Client) ListTimeEntriesBefore(userId int, days int) (*TimeEntryCollection, error) {
	start := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	end := time.Now().Format("2006-01-02")
	filters := fmt.Sprintf(filterTimeEntriesBefore, userId, start, end)
	return c.listTimeEntries(filters)
}

// listTimeEntries is a helper function to get time entries based on filters.
func (c *Client) listTimeEntries(filters string) (*TimeEntryCollection, error) {
	params := url.Values{}
	params.Add("pageSize", "100")
	params.Add("sortBy", "[[\"spent_on\", \"asc\"]]")
	params.Add("filters", filters)
	endpoint := fmt.Sprintf("%stime_entries?%s", c.baseURL, params.Encode())

	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	var collection TimeEntryCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	return &collection, nil
}
