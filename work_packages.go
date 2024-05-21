package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

var (
	// filterWorkPackageAssignedTo is a filter to get open work packages assigned to a specific user.
	filterWorkPackageAssignedTo = "[{\"assigned_to\":{\"operator\":\"=\",\"values\":[\"%d\"]}},{\"status\":{\"operator\":\"o\",\"values\":[]}}]"
)

// WorkPackageCollection represents a collection of work packages.
type WorkPackageCollection struct {
	Embedded struct {
		Elements []WorkPackage `json:"elements"`
	} `json:"_embedded"`
}

// WorkPackage represents a single work package.
type WorkPackage struct {
	Id          int    `json:"id"`
	Type        string `json:"_type"`
	Subject     string `json:"subject"`
	Description struct {
		Raw string `json:"raw"`
	} `json:"description"`
	StartDate      string `json:"startDate"`
	DueDate        string `json:"dueDate"`
	EstimatedTime  string `json:"estimatedTime"`
	SpentTime      string `json:"spentTime"`
	PercentageDone int    `json:"percentageDone"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	Links          struct {
		Project struct {
			Title string `json:"title"`
			Href  string `json:"href"`
		} `json:"project"`
		Status struct {
			Title string `json:"title"`
			Href  string `json:"href"`
		}
	} `json:"_links"`
}

// GetWorkPackage returns a single work package based on its ID.
func (c *Client) GetWorkPackage(workPackageId int) (*WorkPackage, error) {
	endpoint := fmt.Sprintf("%swork_packages/%d", c.baseURL, workPackageId)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	var wp WorkPackage
	if err := json.Unmarshal(body, &wp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	return &wp, nil
}

// ListWorkPackages returns a collection of open work packages assigned to a specific user.
func (c *Client) ListWorkPackages(userId int) (*WorkPackageCollection, error) {
	filters := fmt.Sprintf(filterWorkPackageAssignedTo, userId)
	return c.listWorkPackages(filters)
}

// listTimeEntries is a helper function to get time entries based on filters.
func (c *Client) listWorkPackages(filters string) (*WorkPackageCollection, error) {
	params := url.Values{}
	params.Add("pageSize", "100")
	params.Add("sortBy", "[[\"updated_at\", \"desc\"]]")
	params.Add("groupBy", "status")
	params.Add("select", "*,elements/*,self/status")
	params.Add("filters", filters)
	endpoint := fmt.Sprintf("%swork_packages?%s", c.baseURL, params.Encode())

	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	var collection WorkPackageCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	return &collection, nil
}
