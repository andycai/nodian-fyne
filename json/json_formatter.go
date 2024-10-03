package json

import (
	"bytes"
	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type JSONFormatter struct {
	window fyne.Window
	input  *widget.Entry
	output *widget.Entry
}

func NewJSONFormatter(window fyne.Window) *JSONFormatter {
	formatter := &JSONFormatter{
		window: window,
		input:  widget.NewMultiLineEntry(),
		output: widget.NewMultiLineEntry(),
	}
	formatter.input.SetPlaceHolder("Enter JSON here")
	formatter.output.SetPlaceHolder("Formatted JSON will appear here")
	formatter.output.Disable()

	return formatter
}

func (j *JSONFormatter) CreateUI() fyne.CanvasObject {
	formatPrettyButton := widget.NewButton("Format (Pretty)", func() {
		j.formatJSON(false)
	})
	formatCompactButton := widget.NewButton("Format (Compact)", func() {
		j.formatJSON(true)
	})

	buttons := container.NewHBox(formatPrettyButton, formatCompactButton)
	content := container.NewBorder(nil, buttons, nil, nil,
		container.NewVSplit(j.input, j.output))

	return content
}

func (j *JSONFormatter) formatJSON(compact bool) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(j.input.Text), "", "  ")
	if err != nil {
		j.output.SetText("Error: Invalid JSON")
		return
	}

	if compact {
		var compactOut bytes.Buffer
		err = json.Compact(&compactOut, out.Bytes())
		if err != nil {
			j.output.SetText("Error: Failed to compact JSON")
			return
		}
		j.output.SetText(compactOut.String())
	} else {
		j.output.SetText(out.String())
	}
}
