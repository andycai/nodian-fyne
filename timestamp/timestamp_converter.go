package timestamp

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type TimestampConverter struct {
	window         fyne.Window
	timestampInput *widget.Entry
	dateTimeInput  *widget.Entry
	unitSelect     *widget.Select
	resultLabel    *widget.Label
}

func NewTimestampConverter(window fyne.Window) *TimestampConverter {
	converter := &TimestampConverter{
		window:         window,
		timestampInput: widget.NewEntry(),
		dateTimeInput:  widget.NewEntry(),
		unitSelect:     widget.NewSelect([]string{"Seconds", "Milliseconds"}, nil),
		resultLabel:    widget.NewLabel(""),
	}
	converter.timestampInput.SetPlaceHolder("Enter timestamp")
	converter.dateTimeInput.SetPlaceHolder("YYYY-MM-DD HH:MM:SS")
	converter.unitSelect.SetSelected("Seconds")
	return converter
}

func (t *TimestampConverter) CreateUI() fyne.CanvasObject {
	timestampToDateButton := widget.NewButton("Timestamp to Date", func() {
		t.convertTimestampToDate()
	})
	dateToTimestampButton := widget.NewButton("Date to Timestamp", func() {
		t.convertDateToTimestamp()
	})

	pickTimeButton := widget.NewButton("Pick Time", func() {
		t.showDateTimePicker()
	})

	dateTimeContainer := container.NewBorder(nil, nil, nil, pickTimeButton, t.dateTimeInput)

	content := container.NewVBox(
		widget.NewLabel("Timestamp:"),
		t.timestampInput,
		widget.NewLabel("Date and Time:"),
		dateTimeContainer,
		widget.NewLabel("Unit:"),
		t.unitSelect,
		container.NewHBox(timestampToDateButton, dateToTimestampButton),
		widget.NewLabel("Result:"),
		t.resultLabel,
	)

	return container.NewPadded(content)
}

func (t *TimestampConverter) showDateTimePicker() {
	currentDateTime := time.Now()
	if t.dateTimeInput.Text != "" {
		parsedDateTime, err := time.Parse("2006-01-02 15:04:05", t.dateTimeInput.Text)
		if err == nil {
			currentDateTime = parsedDateTime
		}
	}

	datePicker := dialog.NewCustom("Select Date and Time", "OK", t.createDateTimePickerContent(currentDateTime), t.window)
	datePicker.Resize(fyne.NewSize(300, 400))
	datePicker.Show()
}

func (t *TimestampConverter) createDateTimePickerContent(dateTime time.Time) fyne.CanvasObject {
	year := widget.NewSelect(generateYearList(), nil)
	month := widget.NewSelect(generateMonthList(), nil)
	day := widget.NewSelect(generateDayList(dateTime.Year(), int(dateTime.Month())), nil)
	hour := widget.NewSelect(generateHourList(), nil)
	minute := widget.NewSelect(generateMinuteList(), nil)
	second := widget.NewSelect(generateSecondList(), nil)

	year.SetSelected(fmt.Sprintf("%d", dateTime.Year()))
	month.SetSelected(dateTime.Month().String())
	day.SetSelected(fmt.Sprintf("%d", dateTime.Day()))
	hour.SetSelected(fmt.Sprintf("%02d", dateTime.Hour()))
	minute.SetSelected(fmt.Sprintf("%02d", dateTime.Minute()))
	second.SetSelected(fmt.Sprintf("%02d", dateTime.Second()))

	year.OnChanged = func(value string) {
		y, _ := strconv.Atoi(value)
		m, _ := time.Parse("January", month.Selected)
		day.Options = generateDayList(y, int(m.Month()))
		day.Refresh()
	}

	month.OnChanged = func(value string) {
		y, _ := strconv.Atoi(year.Selected)
		m, _ := time.Parse("January", value)
		day.Options = generateDayList(y, int(m.Month()))
		day.Refresh()
	}

	content := container.NewVBox(
		year,
		month,
		day,
		hour,
		minute,
		second,
		widget.NewButton("Set Date and Time", func() {
			y, _ := strconv.Atoi(year.Selected)
			m, _ := time.Parse("January", month.Selected)
			d, _ := strconv.Atoi(day.Selected)
			h, _ := strconv.Atoi(hour.Selected)
			min, _ := strconv.Atoi(minute.Selected)
			sec, _ := strconv.Atoi(second.Selected)
			selectedDateTime := time.Date(y, m.Month(), d, h, min, sec, 0, time.Local)
			t.dateTimeInput.SetText(selectedDateTime.Format("2006-01-02 15:04:05"))
		}),
	)

	return content
}

func generateYearList() []string {
	currentYear := time.Now().Year()
	years := make([]string, 101)
	for i := 0; i <= 100; i++ {
		years[i] = fmt.Sprintf("%d", currentYear-50+i)
	}
	return years
}

func generateMonthList() []string {
	return []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
}

func generateDayList(year, month int) []string {
	daysInMonth := 31
	switch month {
	case 4, 6, 9, 11:
		daysInMonth = 30
	case 2:
		daysInMonth = 28
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			daysInMonth = 29
		}
	}

	days := make([]string, daysInMonth)
	for i := 1; i <= daysInMonth; i++ {
		days[i-1] = fmt.Sprintf("%d", i)
	}
	return days
}

func generateHourList() []string {
	hours := make([]string, 24)
	for i := 0; i < 24; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}
	return hours
}

func generateMinuteList() []string {
	minutes := make([]string, 60)
	for i := 0; i < 60; i++ {
		minutes[i] = fmt.Sprintf("%02d", i)
	}
	return minutes
}

func generateSecondList() []string {
	return generateMinuteList()
}

func (t *TimestampConverter) convertTimestampToDate() {
	timestamp, err := strconv.ParseInt(t.timestampInput.Text, 10, 64)
	if err != nil {
		t.resultLabel.SetText("Error: Invalid timestamp")
		return
	}

	var tm time.Time
	if t.unitSelect.Selected == "Seconds" {
		tm = time.Unix(timestamp, 0)
	} else {
		tm = time.UnixMilli(timestamp)
	}

	t.dateTimeInput.SetText(tm.Format("2006-01-02 15:04:05"))
	t.resultLabel.SetText(tm.Format("2006-01-02 15:04:05"))
}

func (t *TimestampConverter) convertDateToTimestamp() {
	tm, err := time.Parse("2006-01-02 15:04:05", t.dateTimeInput.Text)
	if err != nil {
		t.resultLabel.SetText("Error: Invalid date-time format")
		return
	}

	var result int64
	if t.unitSelect.Selected == "Seconds" {
		result = tm.Unix()
	} else {
		result = tm.UnixMilli()
	}

	t.resultLabel.SetText(fmt.Sprintf("%d", result))
	t.timestampInput.SetText(fmt.Sprintf("%d", result))
}
