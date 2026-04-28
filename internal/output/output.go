package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format represents the output format
type Format string

const (
	FormatJSON     Format = "json"
	FormatTable    Format = "table"
	FormatMinimal  Format = "minimal"
	FormatTSV      Format = "tsv"
	FormatCSV      Format = "csv"
	FormatYAML     Format = "yaml"
	FormatMarkdown Format = "markdown"
)

var (
	currentFormat Format
	prettyPrint   bool
	quietMode     bool
	writer        io.Writer = os.Stdout
)

// Setup initializes the output formatter
func Setup(format string, pretty, quiet bool) {
	currentFormat = normalizeFormat(format)
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
	case FormatMarkdown:
		return printMarkdown(data)
	default:
		return printJSON(data)
	}
}

// PrintSuccess prints a success message (respects quiet mode)
func PrintSuccess(format string, args ...interface{}) {
	if !quietMode {
		_, _ = fmt.Fprintf(writer, format+"\n", args...)
	}
}

// PrintInfo prints an info message (respects quiet mode)
func PrintInfo(format string, args ...interface{}) {
	if !quietMode {
		_, _ = fmt.Fprintf(writer, format+"\n", args...)
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

	if _, err := fmt.Fprintln(writer, string(output)); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}

func printTable(data interface{}) error {
	w := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()

	v := reflect.ValueOf(data)

	// Handle slice
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			if _, err := fmt.Fprintln(writer, "(no results)"); err != nil {
				return fmt.Errorf("failed to write table output: %w", err)
			}
			return nil
		}

		// Get headers from first element
		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}

		headers := getStructHeaders(first)
		if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
			return fmt.Errorf("failed to write table header: %w", err)
		}
		if _, err := fmt.Fprintln(w, strings.Repeat("-\t", len(headers))); err != nil {
			return fmt.Errorf("failed to write table separator: %w", err)
		}

		// Print rows
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			if _, err := fmt.Fprintln(w, strings.Join(values, "\t")); err != nil {
				return fmt.Errorf("failed to write table row: %w", err)
			}
		}
	} else if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		// Single struct
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		headers := getStructHeaders(v)
		values := getStructValues(v)
		for i, h := range headers {
			if _, err := fmt.Fprintf(w, "%s:\t%s\n", h, values[i]); err != nil {
				return fmt.Errorf("failed to write table field: %w", err)
			}
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
				if _, err := fmt.Fprintln(writer, elem.Field(0).Interface()); err != nil {
					return fmt.Errorf("failed to write minimal output: %w", err)
				}
			}
		}
	} else if v.Kind() == reflect.Struct && v.NumField() > 0 {
		if _, err := fmt.Fprintln(writer, v.Field(0).Interface()); err != nil {
			return fmt.Errorf("failed to write minimal output: %w", err)
		}
	} else {
		if _, err := fmt.Fprintln(writer, data); err != nil {
			return fmt.Errorf("failed to write minimal output: %w", err)
		}
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
		if _, err := fmt.Fprintln(writer, strings.Join(headers, "\t")); err != nil {
			return fmt.Errorf("failed to write TSV header: %w", err)
		}

		// Print rows
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			if _, err := fmt.Fprintln(writer, strings.Join(values, "\t")); err != nil {
				return fmt.Errorf("failed to write TSV row: %w", err)
			}
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
	if _, err := fmt.Fprint(writer, string(out)); err != nil {
		return fmt.Errorf("failed to write YAML output: %w", err)
	}
	return nil
}

func printMarkdown(data interface{}) error {
	headers, rows, ok := markdownRows(data)
	if !ok {
		return printJSON(data)
	}

	if len(headers) == 0 {
		if _, err := fmt.Fprintln(writer, "(no results)"); err != nil {
			return fmt.Errorf("failed to write markdown output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(writer, "| %s |\n", strings.Join(headers, " | ")); err != nil {
		return fmt.Errorf("failed to write markdown header: %w", err)
	}
	separator := make([]string, len(headers))
	for i := range separator {
		separator[i] = "---"
	}
	if _, err := fmt.Fprintf(writer, "| %s |\n", strings.Join(separator, " | ")); err != nil {
		return fmt.Errorf("failed to write markdown separator: %w", err)
	}

	for _, row := range rows {
		if _, err := fmt.Fprintf(writer, "| %s |\n", strings.Join(row, " | ")); err != nil {
			return fmt.Errorf("failed to write markdown row: %w", err)
		}
	}

	return nil
}

func markdownRows(data interface{}) ([]string, [][]string, bool) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, nil, false
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return []string{"Result"}, [][]string{}, true
		}

		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}

		switch first.Kind() {
		case reflect.Struct:
			headers := getStructHeaders(first)
			rows := make([][]string, 0, v.Len())
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				rows = append(rows, escapeMarkdownRow(getStructValues(elem)))
			}
			return escapeMarkdownRow(headers), rows, true
		case reflect.Map:
			headers := getMapHeaders(first)
			rows := make([][]string, 0, v.Len())
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				rows = append(rows, escapeMarkdownRow(getMapValues(elem, headers)))
			}
			return escapeMarkdownRow(headers), rows, true
		default:
			return nil, nil, false
		}
	case reflect.Struct:
		headers := []string{"FIELD", "VALUE"}
		rows := make([][]string, 0, v.NumField())
		structHeaders := getStructHeaders(v)
		structValues := getStructValues(v)
		for i := range structHeaders {
			rows = append(rows, escapeMarkdownRow([]string{structHeaders[i], structValues[i]}))
		}
		return headers, rows, true
	case reflect.Map:
		headers := []string{"FIELD", "VALUE"}
		keys := getMapHeaders(v)
		rows := make([][]string, 0, len(keys))
		values := getMapValues(v, keys)
		for i := range keys {
			rows = append(rows, escapeMarkdownRow([]string{keys[i], values[i]}))
		}
		return headers, rows, true
	default:
		return nil, nil, false
	}
}

func escapeMarkdownRow(values []string) []string {
	escaped := make([]string, len(values))
	for i, value := range values {
		escaped[i] = strings.ReplaceAll(value, "|", "\\|")
	}
	return escaped
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

func getMapHeaders(v reflect.Value) []string {
	keys := v.MapKeys()
	headers := make([]string, 0, len(keys))
	for _, key := range keys {
		headers = append(headers, fmt.Sprintf("%v", key.Interface()))
	}
	sort.Strings(headers)
	return headers
}

func getMapValues(v reflect.Value, headers []string) []string {
	values := make([]string, 0, len(headers))
	for _, header := range headers {
		value := v.MapIndex(reflect.ValueOf(header))
		if value.IsValid() {
			values = append(values, fmt.Sprintf("%v", value.Interface()))
			continue
		}
		values = append(values, "")
	}
	return values
}

func normalizeFormat(format string) Format {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "md", "markdown":
		return FormatMarkdown
	case string(FormatTable):
		return FormatTable
	case string(FormatMinimal):
		return FormatMinimal
	case string(FormatTSV):
		return FormatTSV
	case string(FormatCSV):
		return FormatCSV
	case string(FormatYAML):
		return FormatYAML
	default:
		return FormatJSON
	}
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
