package fdf

// Map of field names (keys) to values
// The value can be a string, Field, OptionInput, or implement fmt.Stringer
type Inputs map[string]interface{}

// OptionInput indicates the value is an option state.
// Option states are used in dropdowns, checkboxes, and radio buttons.
// For PDFtk dump_data_fields, use this for a FieldStateOption value:
//	# pdftk test.pdf dump_data_fields
// 	FieldType: Button
//	FieldName: checkbox1
//	FieldStateOption: Yes
//	FieldStateOption: Off
//
// E.g. it is common to have Yes and Off for checkboxes:
//	OptionInput("Yes") // checked
//	OptionInput("Off") // unchecked
type OptionInput string

// A Field can be used as an input for any value to set additional flags for the input.
type Field struct {
	Hidden   bool
	ReadOnly bool
	Value    interface{}
}
