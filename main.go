package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
)

func main() {
	config, err := ReadConfig()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	tui := NewTui()
	client := NewClient(config.BaseURL, "apikey", config.APIKey)

	workPackages, err := client.ListWorkPackages(config.UserID)
	if err != nil {
		log.Fatalf("error listing work packages: %v", err)
	}
	tui.SetupWorkPackages(client, config.UserID, workPackages)

	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF1 {
			tui.Pages.SwitchToPage("navigation")
			tui.App.SetFocus(tui.WorkPackageList)
			return nil
		}
		if event.Key() == tcell.KeyF2 {
			// The calendar view is updated every time it's accessed.
			tui.CalendarFlex.Clear()

			timeEntries, err := client.ListTimeEntriesBefore(config.UserID, 7)
			if err != nil {
				log.Fatalf("error listing time entries: %v", err)
			}
			tui.SetupCalendar(timeEntries)

			tui.Pages.SwitchToPage("calendar")
			tui.App.SetFocus(tui.CalendarFlex)
			return nil
		}
		return event
	})

	if err := tui.Start(); err != nil {
		panic(err)
	}
}
