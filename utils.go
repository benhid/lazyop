package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Duration struct to hold hours and minutes.
type Duration struct {
	Hours   int
	Minutes int
}

func NewDuration() Duration {
	return Duration{}
}

// Parse parses a duration string.
func Parse(duration string) (*Duration, error) {
	re := regexp.MustCompile(`(\d+h)?(\d+m)?`)
	matches := re.FindStringSubmatch(duration)
	if matches == nil {
		return nil, fmt.Errorf("invalid duration format")
	}
	d := NewDuration()
	for _, match := range matches[1:] {
		if match == "" {
			continue
		}
		switch {
		case strings.HasSuffix(match, "h"):
			h, err := strconv.Atoi(strings.TrimSuffix(match, "h"))
			if err != nil {
				return nil, err
			}
			d.Hours = h
		case strings.HasSuffix(match, "m"):
			m, err := strconv.Atoi(strings.TrimSuffix(match, "m"))
			if err != nil {
				return nil, err
			}
			d.Minutes = m
		}
	}
	return &d, nil
}

// ParseIso8601 parses a duration string in ISO 8601 format.
func ParseIso8601(duration string) (*Duration, error) {
	re := regexp.MustCompile(`PT(\d+H)?(\d+M)?`)
	matches := re.FindStringSubmatch(duration)
	if matches == nil {
		return nil, fmt.Errorf("invalid duration format: %s", duration)
	}
	d := NewDuration()
	for _, match := range matches[1:] {
		if match == "" {
			continue
		}
		switch {
		case strings.HasSuffix(match, "H"):
			h, err := strconv.Atoi(strings.TrimSuffix(match, "H"))
			if err != nil {
				return nil, err
			}
			d.Hours = h
		case strings.HasSuffix(match, "M"):
			m, err := strconv.Atoi(strings.TrimSuffix(match, "M"))
			if err != nil {
				return nil, err
			}
			d.Minutes = m
		}
	}
	return &d, nil
}

// Add adds another duration to the current duration.
func (d *Duration) Add(other *Duration) {
	d.Hours += other.Hours
	d.Minutes += other.Minutes
	if d.Minutes >= 60 {
		d.Hours++
		d.Minutes -= 60
	}
}

// Adds hours to the current duration.
func (d *Duration) Adds(other *[]Duration) {
	for _, other := range *other {
		d.Add(&other)
	}
}

// ToIso8601String returns the duration as an ISO 8601 string.
func (d *Duration) ToIso8601String() string {
	var isoDuration []string
	isoDuration = append(isoDuration, "PT")
	if d.Hours > 0 {
		isoDuration = append(isoDuration, fmt.Sprintf("%dH", d.Hours))
	}
	if d.Minutes > 0 {
		isoDuration = append(isoDuration, fmt.Sprintf("%dM", d.Minutes))
	}
	return strings.Join(isoDuration, "")
}

// ToString returns the duration as a human-readable string in the format "HhMm".
func (d *Duration) ToString() string {
	var humanReadable []string
	humanReadable = append(humanReadable, fmt.Sprintf("%dh", d.Hours))
	if d.Minutes > 0 {
		humanReadable = append(humanReadable, fmt.Sprintf("%dm", d.Minutes))
	}
	return strings.Join(humanReadable, "")
}
