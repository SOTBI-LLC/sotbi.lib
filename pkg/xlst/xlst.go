package xlst

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aymerick/raymond"
	"github.com/tealeg/xlsx/v3"

	"github.com/COTBU/sotbi.lib/pkg/times"
)

var (
	rgx         = regexp.MustCompile(`{{\s*(\w+)\.\w+\s*}}`)
	rangeRgx    = regexp.MustCompile(`{{\s*range\s+(\w+)\s*}}`)
	rangeEndRgx = regexp.MustCompile(`{{\s*end\s*}}`)
)

// Xlst Represents template struct.
type Xlst struct {
	file   *xlsx.File
	report *xlsx.File
}

// Options for render has only one property WrapTextInAllCells for wrapping text.
type Options struct {
	WrapTextInAllCells bool
}

// New creates new Xlst struct and returns pointer to it.
func New() *Xlst {
	return &Xlst{}
}

// NewFromBinary creates new Xlst struct puts binary template into and returns pointer to it.
func NewFromBinary(content []byte) (*Xlst, error) {
	file, err := xlsx.OpenBinary(content)
	if err != nil {
		return nil, err
	}

	res := &Xlst{file: file}

	return res, nil
}

// Render renders report and stores it in a struct.
func (m *Xlst) Render(in interface{}) error {
	return m.RenderWithOptions(in, nil)
}

// RenderWithOptions renders report with options provided and stores it in a struct.
func (m *Xlst) RenderWithOptions(in interface{}, options *Options) error {
	if options == nil {
		options = new(Options)
	}

	report := xlsx.NewFile()

	for sheetIndex, sheet := range m.file.Sheets {
		ctx := getCtx(in, sheetIndex)

		_, err := report.AddSheet(sheet.Name)
		if err != nil {
			return err
		}

		cloneSheet(sheet, report.Sheets[sheetIndex])

		rows := make([]*xlsx.Row, sheet.MaxRow)

		for i := 0; i < sheet.MaxRow; i++ {
			row, err := sheet.Row(i)
			if err != nil {
				return err
			}

			rows[i] = row
		}

		if err = renderRows(report.Sheets[sheetIndex], rows, ctx, options); err != nil {
			return err
		}

		sheet.Cols.ForEach(func(_ int, col *xlsx.Col) {
			report.Sheets[sheetIndex].Cols.Add(col)
		})
	}

	m.report = report

	return nil
}

// ReadTemplate reads template from disk and stores it in a struct.
func (m *Xlst) ReadTemplate(path string) error {
	file, err := xlsx.OpenFile(path)
	if err != nil {
		return err
	}

	m.file = file

	return nil
}

// Save saves generated report to disk.
func (m *Xlst) Save(path string) error {
	if m.report == nil {
		return errors.New("report was not generated")
	}

	return m.report.Save(path)
}

// Write writes generated report to provided writer.
func (m *Xlst) Write(writer io.Writer) error {
	if m.report == nil {
		return errors.New("report was not generated")
	}

	return m.report.Write(writer)
}

func renderRows(
	sheet *xlsx.Sheet,
	rows []*xlsx.Row,
	ctx map[string]interface{},
	options *Options,
) error {
	for ri := 0; ri < len(rows); ri++ {
		row := rows[ri]

		if rangeProp := getRangeProp(row); rangeProp != "" {
			ri++

			rangeEndIndex := getRangeEndIndex(rows[ri:])
			if rangeEndIndex == -1 {
				return fmt.Errorf("end of range %q not found", rangeProp)
			}

			rangeEndIndex += ri

			rangeCtx := getRangeCtx(ctx, rangeProp)
			if rangeCtx == nil {
				return fmt.Errorf("not expected context property for range %q", rangeProp)
			}

			for idx := range rangeCtx {
				localCtx := mergeCtx(rangeCtx[idx], ctx)

				err := renderRows(sheet, rows[ri:rangeEndIndex], localCtx, options)
				if err != nil {
					return err
				}
			}

			ri = rangeEndIndex

			continue
		}

		prop := getListProp(row)
		if prop == "" {
			newRow := sheet.AddRow()
			cloneRow(row, newRow, options)

			if err := renderRow(newRow, ctx); err != nil {
				return err
			}

			continue
		}

		if !isArray(ctx, prop) {
			newRow := sheet.AddRow()
			cloneRow(row, newRow, options)

			err := renderRow(newRow, ctx)
			if err != nil {
				return err
			}

			continue
		}

		arr := reflect.ValueOf(ctx[prop])
		arrBackup := ctx[prop]

		for i := 0; i < arr.Len(); i++ {
			newRow := sheet.AddRow()
			cloneRow(row, newRow, options)

			ctx[prop] = arr.Index(i).Interface()

			err := renderRow(newRow, ctx)
			if err != nil {
				return err
			}
		}

		ctx[prop] = arrBackup
	}

	return nil
}

func cloneCell(from, to *xlsx.Cell, options *Options) {
	to.Value = from.Value
	style := from.GetStyle()

	if options.WrapTextInAllCells {
		style.Alignment.WrapText = true
	}

	to.SetStyle(style)
	to.HMerge = from.HMerge
	to.VMerge = from.VMerge
	to.Hidden = from.Hidden
	to.NumFmt = from.NumFmt
}

func cloneRow(from, to *xlsx.Row, options *Options) {
	if from.GetHeight() != 0 {
		to.SetHeight(from.GetHeight())
	}

	if err := from.ForEachCell(func(cell *xlsx.Cell) error {
		newCell := to.AddCell()
		cloneCell(cell, newCell, options)

		return nil
	}); err != nil {
		return
	}
}

// Вставка формул:
// Excel не даёт просто так вставить текст формулы в шаблон - он сразу её считает,
// и наша система не может её распарсить.
// Нужно:
// 1. Сменить формат клетки на текстовый.
// 2. Формула должна быть в виде = , потом пробел, потом формула.
// 3. Вставляем это просто как текст.
// 4. Меняем формат клетки на нужный.
// Возможно, есть другой способ, но пока получалось только так.
func renderCell(cell *xlsx.Cell, ctx interface{}) error {
	if strings.HasPrefix(cell.String(), "=") {
		formula := strings.TrimPrefix(cell.Value, "= ")
		cell.SetStringFormula(formula) // метод устанавливает для клетки значение формулы
		// cell.SetFormat("#0.00")           //уточнить про этот формат

		return nil
	}

	fm := cell.GetNumberFormat()

	// cell value processing
	cellValue := strings.ReplaceAll(cell.Value, "{{", "{{{")
	cellValue = strings.ReplaceAll(cellValue, "}}", "}}}")

	template, err := raymond.Parse(cellValue)
	if err != nil {
		return err
	}

	out, err := template.Exec(ctx)
	if err != nil {
		return err
	}

	dt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", out)
	if err == nil {
		location := times.GetMoscowLocation()
		opt := xlsx.DateTimeOptions{Location: location, ExcelTimeFormat: "dd.mm.yyyy"}
		cell.SetDateWithOptions(dt, opt)

		return nil
	}

	iVal, err := strconv.ParseInt(out, 10, 64)
	if err == nil {
		if iVal > 90000000000 {
			cell.SetString(strconv.FormatInt(iVal, 10))
		} else {
			cell.SetInt64(iVal)
			cell.SetFormat(fm)
		}

		return nil
	}

	fVal, err := strconv.ParseFloat(out, 64)
	if err == nil && len(out) != 20 { // https://ourzoo.online:8443/browse/BH-291
		if fVal > 90000000000 {
			cell.SetString(strconv.FormatFloat(fVal, 'f', 0, 64))
		} else {
			// cell.SetFloatWithFormat(fVal, "#0")
			cell.SetFloat(fVal)
			cell.SetFormat(fm)
		}

		return nil
	}

	if out == "<nil>" {
		cell.Value = ""

		return nil
	}

	cell.Value = out

	return nil
}

func cloneSheet(from, to *xlsx.Sheet) {
	if from.Cols == nil {
		return
	}

	from.Cols.ForEach(func(_ int, col *xlsx.Col) {
		newCol := &xlsx.Col{}
		style := col.GetStyle()
		newCol.SetStyle(style)
		newCol.Width = col.Width
		newCol.Hidden = col.Hidden
		newCol.Collapsed = col.Collapsed
		newCol.Min = col.Min
		newCol.Max = col.Max
		to.Cols.Add(newCol)
	})
}

func getCtx(in interface{}, i int) map[string]interface{} {
	if ctx, ok := in.(map[string]interface{}); ok {
		return ctx
	}

	if ctxSlice, ok := in.([]interface{}); ok {
		if len(ctxSlice) > i {
			_ctx := ctxSlice[i]
			if ctx, ok := _ctx.(map[string]interface{}); ok {
				return ctx
			}
		}

		return nil
	}

	return nil
}

func getRangeCtx(ctx map[string]interface{}, prop string) []map[string]interface{} {
	val, ok := ctx[prop]
	if !ok {
		return nil
	}

	if propCtx, ok := val.([]map[string]interface{}); ok {
		return propCtx
	}

	return nil
}

func mergeCtx(local, global map[string]interface{}) map[string]interface{} {
	ctx := make(map[string]interface{})

	for k, v := range global {
		ctx[k] = v
	}

	for k, v := range local {
		ctx[k] = v
	}

	return ctx
}

func isArray(in map[string]interface{}, prop string) bool {
	val, ok := in[prop]
	if !ok {
		return false
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func getListProp(in *xlsx.Row) string {
	matchedProp := ""

	if err := in.ForEachCell(func(cell *xlsx.Cell) error {
		if cell.Value == "" || matchedProp != "" {
			return nil
		}

		if match := rgx.FindAllStringSubmatch(cell.Value, -1); match != nil {
			matchedProp = match[0][1]

			return nil
		}

		return nil
	}); err != nil {
		return ""
	}

	return matchedProp
}

func getRangeProp(in *xlsx.Row) string {
	cell := in.GetCell(0)

	match := rangeRgx.FindAllStringSubmatch(cell.Value, -1)
	if match != nil {
		return match[0][1]
	}

	return ""
}

func getRangeEndIndex(rows []*xlsx.Row) int {
	var nesting int

	for idx := 0; idx < len(rows); idx++ {
		cell := rows[idx].GetCell(0)

		if cell.Value == "" {
			continue
		}

		if rangeEndRgx.MatchString(cell.Value) {
			if nesting == 0 {
				return idx
			}

			nesting--

			continue
		}

		if rangeRgx.MatchString(cell.Value) {
			nesting++
		}
	}

	return -1
}

func renderRow(in *xlsx.Row, ctx interface{}) error {
	return in.ForEachCell(func(cell *xlsx.Cell) error {
		return renderCell(cell, ctx)
	})
}
