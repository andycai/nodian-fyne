package timestamp

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type TimestampConverter struct {
	window fyne.Window
	input  *widget.Entry
	output *widget.Entry
	unit   *widget.Select
}

func NewTimestampConverter(window fyne.Window) *TimestampConverter {
	converter := &TimestampConverter{
		window: window,
		input:  widget.NewEntry(),
		output: widget.NewEntry(),
		unit:   widget.NewSelect([]string{"Seconds", "Milliseconds"}, nil),
	}
	converter.input.SetPlaceHolder("Enter timestamp or date-time")
	converter.output.SetPlaceHolder("Conversion result will appear here")
	converter.output.Disable()
	converter.unit.SetSelected("Seconds")

	return converter
}

func (t *TimestampConverter) CreateUI() fyne.CanvasObject {
	convertButton := widget.NewButton("Convert", func() {
		t.convert()
	})

	content := container.NewVBox(
		widget.NewLabel("Timestamp/Date-Time:"),
		t.input,
		widget.NewLabel("Unit:"),
		t.unit,
		convertButton,
		widget.NewLabel("Result:"),
		t.output,
	)

	return container.NewPadded(content)
}

func (t *TimestampConverter) convert() {
	input := t.input.Text
	unit := t.unit.Selected

	// Try to parse input as timestamp
	timestamp, err := strconv.ParseInt(input, 10, 64)
	if err == nil {
		// Input is a timestamp, convert to date-time
		var tm time.Time
		if unit == "Seconds" {
			tm = time.Unix(timestamp, 0)
		} else {
			tm = time.UnixMilli(timestamp)
		}
		t.output.SetText(tm.Format(time.RFC3339))
	} else {
		// Input might be a date-time, try to parse and convert to timestamp
		tm, err := time.Parse(time.RFC3339, input)
		if err != nil {
			t.output.SetText("Error: Invalid input format")
			return
		}
		if unit == "Seconds" {
			t.output.SetText(fmt.Sprintf("%d", tm.Unix()))
		} else {
			t.output.SetText(fmt.Sprintf("%d", tm.UnixMilli()))
		}
	}
}
