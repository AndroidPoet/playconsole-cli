package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format represents the output format
type Format string

const (
	FormatJSON    Format = "json"
	FormatTable   Format = "table"
	FormatMinimal Format = "minimal"
	FormatTSV     Format = "tsv"
	FormatCSV     Format = "csv"
	FormatYAML    Format = "yaml"
)

var (
	currentFormat Format
	prettyPrint   bool
	quietMode     bool
	writer        io.Writer = os.Stdout
)

// Setup initializes the output formatter
func Setup(format string, pretty, quiet bool) {
	currentFormat = Format(format)
	prettyPrint = pretty
	quietMode = quiet
}

// SetWriter sets the output writer (for testing)
func SetWriter(w io.Writer) {
	writer = w
}

// Print outputs data in the configured format
func Print(data interface{}) error {
	switch currentFormat {
	case FormatJSON:
		return printJSON(data)
	case FormatTable:
		return printTable(data)
	case FormatMinimal:
		return printMinimal(data)
	case FormatTSV:
		return printTSV(data)
	case FormatCSV:
		return printCSV(data)
	case FormatYAML:
		return printYAML(data)
	default:
		return printJSON(data)
	}
}

// PrintSuccess prints a success message (respects quiet mode)
func PrintSuccess(format string, args ...interface{}) {
	if !quietMode {
		fmt.Fprintf(writer, format+"\n", args...)
	}
}

// PrintInfo prints an info message (respects quiet mode)
func PrintInfo(format string, args ...interface{}) {
	if !quietMode {
		fmt.Fprintf(writer, format+"\n", args...)
	}
}

// PrintError prints an error message (always shown)
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	if !quietMode {
		fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
	}
}

func printJSON(data interface{}) error {
	var output []byte
	var err error

	if prettyPrint {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Fprintln(writer, string(output))
	return nil
}

func printTable(data interface{}) error {
	w := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	v := reflect.ValueOf(data)

	// Handle slice
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			fmt.Fprintln(writer, "(no results)")
			return nil
		}

		// Get headers from first element
		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}

		headers := getStructHeaders(first)
		fmt.Fprintln(w, strings.Join(headers, "\t"))
		fmt.Fprintln(w, strings.Repeat("-\t", len(headers)))

		// Print rows
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			fmt.Fprintln(w, strings.Join(values, "\t"))
		}
	} else if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		// Single struct
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		headers := getStructHeaders(v)
		values := getStructValues(v)
		for i, h := range headers {
			fmt.Fprintf(w, "%s:\t%s\n", h, values[i])
		}
	} else {
		// Fallback to JSON
		return printJSON(data)
	}

	return nil
}

func printMinimal(data interface{}) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			// Print first field value
			if elem.Kind() == reflect.Struct && elem.NumField() > 0 {
				fmt.Fprintln(writer, elem.Field(0).Interface())
			}
		}
	} else if v.Kind() == reflect.Struct && v.NumField() > 0 {
		fmt.Fprintln(writer, v.Field(0).Interface())
	} else {
		fmt.Fprintln(writer, data)
	}

	return nil
}

func printTSV(data interface{}) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return nil
		}

		// Print header
		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}
		headers := getStructHeaders(first)
		fmt.Fprintln(writer, strings.Join(headers, "\t"))

		// Print rows
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			fmt.Fprintln(writer, strings.Join(values, "\t"))
		}
	} else {
		return printJSON(data)
	}

	return nil
}

func printCSV(data interface{}) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return nil
		}

		// Print header
		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}
		headers := getStructHeaders(first)
		if err := w.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}

		// Print rows
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			if err := w.Write(values); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	} else {
		return printJSON(data)
	}

	return nil
}

func printYAML(data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	fmt.Fprint(writer, string(out))
	return nil
}

func getStructHeaders(v reflect.Value) []string {
	t := v.Type()
	headers := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Use json tag if available
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			headers = append(headers, strings.ToUpper(parts[0]))
		} else {
			headers = append(headers, strings.ToUpper(field.Name))
		}
	}
	return headers
}

func getStructValues(v reflect.Value) []string {
	values := make([]string, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		values = append(values, fmt.Sprintf("%v", field.Interface()))
	}
	return values
}

// Result wraps a successful operation result
type Result struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewResult creates a new result
func NewResult(success bool, message string, data interface{}) *Result {
	return &Result{
		Success: success,
		Message: message,
		Data:    data,
	}
}

// PrintResult prints an operation result
func PrintResult(success bool, message string, data interface{}) error {
	return Print(NewResult(success, message, data))
}
