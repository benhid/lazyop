package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	help = "<[yellow]N[green]>ew Entry <[yellow]E[green]>dit Entry <[red]D[green]>elete Entry <[yellow]ESC[green]> Return to the list"
)

type Tui struct {
	App   *tview.Application
	Pages *tview.Pages

	// WorkPackageList view on the left side.
	WorkPackageList *tview.List

	// WorkPackageTextView view on the top right side.
	WorkPackageTextView *tview.TextView

	// TimeEntriesFrame view on the bottom right side.
	TimeEntriesFrame *tview.Frame
	TimeEntriesTable *tview.Table

	// CalendarFlex page.
	CalendarFlex *tview.Flex

	// wp is the currently selected work package.
	wp *WorkPackage
}

func NewTui() *Tui {
	tui := tview.NewApplication()

	workPackageList := tview.NewList()
	workPackageList.ShowSecondaryText(false)
	workPackageList.SetBorder(true).SetTitle("Work Packages")

	workPackageTextView := tview.NewTextView().SetDynamicColors(true)
	workPackageTextView.SetBorder(true).SetTitle("Work Package Details")

	timeEntriesTable := tview.NewTable()
	timeEntriesTable.SetBorders(false)
	timeEntriesTable.SetSelectable(true, false)

	timeEntriesFrame := tview.NewFrame(timeEntriesTable)
	timeEntriesFrame.AddText(help, false, tview.AlignCenter, tview.Styles.PrimaryTextColor)
	timeEntriesFrame.SetBorder(true).SetTitle("Time Entries")

	flex := tview.NewFlex().
		AddItem(workPackageList, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(workPackageTextView, 0, 1, false).
			AddItem(timeEntriesFrame, 0, 3, false), 0, 2, false)

	calendarFlex := tview.NewFlex()

	pages := tview.NewPages().
		AddPage("navigation", flex, true, true).
		AddPage("calendar", calendarFlex, true, false)

	// Navigation.
	workPackageList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			tui.SetFocus(timeEntriesTable)
		}
		return event
	})

	timeEntriesFrame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			tui.SetFocus(workPackageList)
		}
		return event
	})

	return &Tui{
		App:                 tui,
		Pages:               pages,
		WorkPackageList:     workPackageList,
		WorkPackageTextView: workPackageTextView,
		TimeEntriesFrame:    timeEntriesFrame,
		TimeEntriesTable:    timeEntriesTable,
		CalendarFlex:        calendarFlex,
		wp:                  nil,
	}
}

func (tui *Tui) SetupWorkPackages(client *Client, userId int, workPackages *WorkPackageCollection) {
	tui.WorkPackageList.SetChangedFunc(func(idx int, mainText string, secondaryText string, shortcut rune) {
		// A work package was selected. Show its details.
		tui.WorkPackageTextView.Clear()
		tui.TimeEntriesTable.Clear()

		wp := workPackages.Embedded.Elements[idx]
		details, err := client.GetWorkPackage(wp.Id)
		if err != nil {
			tui.ShowError(err)
			return
		}
		tui.wp = details
		tui.SetupWorkPackage(details)

		timeEntries, err := client.ListTimeEntries(wp.Id)
		if err != nil {
			tui.ShowError(err)
			return
		}
		tui.SetupTimeEntries(timeEntries, wp.Id)

		tui.TimeEntriesTable.
			SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyRune {
					switch event.Rune() {
					case 'n':
						tui.showNewTimeEntryForm(client, userId, wp.Id, idx)
					case 'e':
						tui.showEditTimeEntryForm(client, idx)
					case 'd':
						tui.showDeleteTimeEntryForm(client, idx)
					}
				}
				return event
			})

		total := NewDuration()
		for _, te := range timeEntries.Embedded.Elements {
			hours, err := ParseIso8601(te.Hours)
			if err != nil {
				return
			}
			total.Add(hours)
		}

		tui.TimeEntriesFrame.Clear()
		tui.TimeEntriesFrame.AddText(help, false, tview.AlignCenter, tview.Styles.PrimaryTextColor)
		// We also need to reset the help text, because it's cleared by the `Clear` call above.
		tui.TimeEntriesFrame.AddText(fmt.Sprintf("Total: %s", total.ToString()), false, tview.AlignCenter, tcell.ColorYellow)
	})
	for _, wp := range workPackages.Embedded.Elements {
		title := fmt.Sprintf("[green]%s[white]: %s", wp.Links.Project.Title, wp.Subject)
		tui.WorkPackageList.AddItem(title, "", 0, nil)
	}
}

func (tui *Tui) SetupWorkPackage(wp *WorkPackage) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[green]ID[white]: %d\n", wp.Id))
	builder.WriteString(fmt.Sprintf("[green]Type[white]: %s\n", wp.Type))
	builder.WriteString(fmt.Sprintf("[green]Status[white]: %s\n", wp.Links.Status.Title))
	builder.WriteString(fmt.Sprintf("[green]Subject[white]: %s\n", wp.Subject))
	estimatedTime, err := ParseIso8601(wp.EstimatedTime)
	if err == nil {
		builder.WriteString(fmt.Sprintf("[green]Estimated Time[white]: %s\n", estimatedTime.ToString()))
	}
	if wp.Description.Raw != "" {
		builder.WriteString(fmt.Sprintf("[green]Description[white]: %s\n", wp.Description.Raw))
	}
	tui.WorkPackageTextView.SetText(builder.String())
}

func (tui *Tui) SetupTimeEntries(timeEntries *TimeEntryCollection, workPackageId int) {
	headers := []string{"Work Package", "ID", "Duration", "Date", "Comment"}
	for i, header := range headers {
		tui.TimeEntriesTable.SetCell(0, i, tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false))
	}
	for i, te := range timeEntries.Embedded.Elements {
		hours, err := ParseIso8601(te.Hours)
		if err != nil {
			tui.ShowError(err)
			return
		}
		cellColor := color(te.Comment.Raw)
		tui.TimeEntriesTable.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("%d", workPackageId)).SetTextColor(cellColor))
		tui.TimeEntriesTable.SetCell(i+1, 1, tview.NewTableCell(fmt.Sprintf("%d", te.Id)).SetTextColor(cellColor))
		tui.TimeEntriesTable.SetCell(i+1, 2, tview.NewTableCell(hours.ToString()).SetTextColor(cellColor))
		tui.TimeEntriesTable.SetCell(i+1, 3, tview.NewTableCell(te.Date).SetTextColor(cellColor))
		tui.TimeEntriesTable.SetCell(i+1, 4, tview.NewTableCell(te.Comment.Raw).SetExpansion(1).SetTextColor(cellColor))
	}
	tui.TimeEntriesTable.ScrollToBeginning()
}

func (tui *Tui) SetupCalendar(timeEntries *TimeEntryCollection) {
	timeEntriesDate := make(map[string][]TimeEntry)
	for _, te := range timeEntries.Embedded.Elements {
		timeEntriesDate[te.Date] = append(timeEntriesDate[te.Date], te)
	}
	ordered := make([]string, 0, len(timeEntriesDate))
	for date := range timeEntriesDate {
		ordered = append(ordered, date)
	}
	sort.Strings(ordered)
	for _, date := range ordered {
		tes := timeEntriesDate[date]
		table := tview.NewTable()
		table.SetBorder(true).SetTitle(date)
		headers := []string{"ID", "Duration", "Comment"}
		for i, header := range headers {
			table.SetCell(0, i, tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false))
		}
		for i, te := range tes {
			hours, err := ParseIso8601(te.Hours)
			if err != nil {
				tui.ShowError(err)
				return
			}
			cellColor := color(te.Comment.Raw)
			table.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("%d", te.Id)).SetTextColor(cellColor))
			table.SetCell(i+1, 1, tview.NewTableCell(hours.ToString()).SetTextColor(cellColor))
			table.SetCell(i+1, 2, tview.NewTableCell(te.Comment.Raw).SetExpansion(1).SetTextColor(cellColor))
		}

		total := NewDuration()
		for _, te := range tes {
			hours, err := ParseIso8601(te.Hours)
			if err != nil {
				return
			}
			total.Add(hours)
		}

		frame := tview.NewFrame(table).
			AddText(fmt.Sprintf("Total: %s", total.ToString()), false, tview.AlignCenter, tcell.ColorYellow).
			SetBorders(0, 0, 0, 0, 0, 0)
		tui.CalendarFlex.AddItem(frame, 0, 1, false)
	}
}

func (tui *Tui) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func (tui *Tui) ShowError(err error) {
	modal := tview.NewModal().
		SetText(err.Error()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			tui.Pages.RemovePage("error")
		})
	tui.Pages.AddPage("error", modal, true, true)
}

func (tui *Tui) Start() error {
	return tui.App.SetRoot(tui.Pages, true).EnableMouse(true).Run()
}

func (tui *Tui) showNewTimeEntryForm(client *Client, userId int, workPackageId int, workPackageIndex int) {
	form := tview.NewForm()
	form.AddInputField("Hours", "1h30m", 0, nil, nil).
		AddInputField("Comment", "", 0, nil, nil).
		AddInputField("Spent on", time.Now().Format("2006-01-02"), 0, nil, nil).
		AddButton("Save", func() {
			hours, err := Parse(form.GetFormItem(0).(*tview.InputField).GetText())
			if err != nil {
				tui.ShowError(fmt.Errorf("invalid duration input: %v", err))
				return
			}

			comment := form.GetFormItem(1).(*tview.InputField).GetText()
			spentOn := form.GetFormItem(2).(*tview.InputField).GetText()

			te := &TimeEntryRequest{}
			te.Comment.Raw = comment
			te.Hours = hours.ToIso8601String()
			te.Date = spentOn
			te.Links.WorkPackage.Href = fmt.Sprintf("/api/v3/work_packages/%d", workPackageId)
			te.User.Href = fmt.Sprintf("/api/v3/users/%d", userId)
			te.Activity.Href = "/api/v3/time_entries/activities/1"

			if err := client.CreateTimeEntry(te); err != nil {
				tui.ShowError(err)
				return
			}

			tui.Pages.HidePage("newTimeEntryForm")
			tui.App.SetFocus(tui.TimeEntriesTable)
			// `SetCurrentItem` doesn't trigger a `change` event if the item is already selected,
			// so we need to switch to another item first. This is a workaround.
			tui.WorkPackageList.SetCurrentItem(workPackageIndex + 1)
			tui.WorkPackageList.SetCurrentItem(workPackageIndex)
		}).
		AddButton("Quit", func() {
			tui.Pages.HidePage("newTimeEntryForm")
			tui.App.SetFocus(tui.TimeEntriesTable)
		})

	form.SetBorder(true).SetTitle("Log Time").SetTitleAlign(tview.AlignCenter)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetTitleColor(tcell.ColorYellow)
	form.SetCancelFunc(func() {
		tui.Pages.HidePage("newTimeEntryForm")
		tui.App.SetFocus(tui.TimeEntriesTable)
	})

	tui.Pages.AddPage("newTimeEntryForm", tui.Modal(form, 45, 11), true, true)
}

func (tui *Tui) showEditTimeEntryForm(client *Client, workPackageIndex int) {
	row, _ := tui.TimeEntriesTable.GetSelection()
	if row < 0 {
		return
	}

	teId := tui.TimeEntriesTable.GetCell(row, 1).Text
	if teId == "" {
		return
	}

	timeEntryId, _ := strconv.Atoi(teId)
	timeEntryHours := tui.TimeEntriesTable.GetCell(row, 2).Text
	timeEntrySpentOn := tui.TimeEntriesTable.GetCell(row, 3).Text
	timeEntryComment := tui.TimeEntriesTable.GetCell(row, 4).Text

	form := tview.NewForm()
	form.AddInputField("Hours", timeEntryHours, 0, nil, nil).
		AddInputField("Comment", timeEntryComment, 0, nil, nil).
		AddInputField("Spent on", timeEntrySpentOn, 0, nil, nil).
		AddButton("Save changes", func() {
			hours, err := Parse(form.GetFormItem(0).(*tview.InputField).GetText())
			if err != nil {
				tui.ShowError(fmt.Errorf("invalid duration input: %v", err))
				return
			}

			comment := form.GetFormItem(1).(*tview.InputField).GetText()
			spentOn := form.GetFormItem(2).(*tview.InputField).GetText()

			if err := client.UpdateTimeEntryDuration(timeEntryId, hours.ToIso8601String(), comment, spentOn); err != nil {
				tui.ShowError(err)
				return
			}

			tui.Pages.HidePage("editTimeEntryForm")
			tui.App.SetFocus(tui.TimeEntriesTable)
			// `SetCurrentItem` doesn't trigger a `change` event if the item is already selected,
			// so we need to switch to another item first. This is a workaround.
			tui.WorkPackageList.SetCurrentItem(workPackageIndex + 1)
			tui.WorkPackageList.SetCurrentItem(workPackageIndex)
		}).
		AddButton("Quit", func() {
			tui.Pages.HidePage("editTimeEntryForm")
			tui.App.SetFocus(tui.TimeEntriesTable)
		})

	form.SetBorder(true).SetTitle(fmt.Sprintf("Edit Time Entry %d", timeEntryId)).SetTitleAlign(tview.AlignCenter)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetTitleColor(tcell.ColorYellow)
	form.SetCancelFunc(func() {
		tui.Pages.HidePage("editTimeEntryForm")
		tui.App.SetFocus(tui.TimeEntriesTable)
	})

	tui.Pages.AddPage("editTimeEntryForm", tui.Modal(form, 45, 11), true, true)
}

func (tui *Tui) showDeleteTimeEntryForm(client *Client, workPackageIndex int) {
	row, _ := tui.TimeEntriesTable.GetSelection()
	if row < 0 {
		return
	}

	teId := tui.TimeEntriesTable.GetCell(row, 1).Text
	if teId == "" {
		return
	}

	timeEntryId, _ := strconv.Atoi(teId)

	form := tview.NewForm()
	form.AddTextView("", "Are you sure you want to delete this time entry?", 0, 0, false, true).
		AddButton("Yes", func() {
			if err := client.DeleteTimeEntry(timeEntryId); err != nil {
				tui.ShowError(err)
				return
			}
			tui.Pages.HidePage("deleteTimeEntryForm")
			tui.App.SetFocus(tui.TimeEntriesTable)
			// `SetCurrentItem` doesn't trigger a `change` event if the item is already selected,
			// so we need to switch to another item first. This is a workaround.
			tui.WorkPackageList.SetCurrentItem(workPackageIndex + 1)
			tui.WorkPackageList.SetCurrentItem(workPackageIndex)
		}).
		AddButton("Quit", func() {
			tui.Pages.HidePage("deleteTimeEntryForm")
			tui.App.SetFocus(tui.TimeEntriesTable)
		})

	form.SetBorder(true).SetTitle(fmt.Sprintf("Delete Time Entry %d", timeEntryId)).SetTitleAlign(tview.AlignCenter)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetTitleColor(tcell.ColorYellow)
	form.SetCancelFunc(func() {
		tui.Pages.HidePage("deleteTimeEntryForm")
		tui.App.SetFocus(tui.TimeEntriesTable)
	})

	tui.Pages.AddPage("deleteTimeEntryForm", tui.Modal(form, 45, 11), true, true)
}

func color(comment string) tcell.Color {
	if comment == "" {
		return tcell.ColorRed
	}
	return tcell.ColorWhite
}
