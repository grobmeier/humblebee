// Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package guiapp

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grobmeier/humblebee/internal/paths"
)

func (a *App) ExportWorktimeByMonthReport(req ReportRequest) (string, error) {
	report, err := a.GetWorktimeByMonthReport(req)
	if err != nil {
		return "", err
	}
	labels := reportExportLabelsFor(req.Language)
	rows := [][]string{
		{labels.Project, labels.Task, labels.Date, labels.Start, labels.End, labels.Duration, labels.Description},
	}
	for _, row := range report.Rows {
		rows = append(rows, []string{
			row.ProjectName,
			row.TaskName,
			row.Date,
			row.StartTime,
			row.EndTime,
			row.Duration,
			row.Description,
		})
	}
	rows = append(rows, []string{"", "", "", "", labels.Total, report.TotalDuration, ""})

	path, err := reportExportPath("worktime-by-month", req)
	if err != nil {
		return "", err
	}
	if err := writeSimpleXLSX(path, labels.ReportSheet, rows); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ExportWorktimeGroupedByProjectReport(req ReportRequest) (string, error) {
	report, err := a.GetWorktimeGroupedByProjectReport(req)
	if err != nil {
		return "", err
	}
	labels := reportExportLabelsFor(req.Language)
	rows := [][]string{{labels.Project, labels.Task, labels.Date, labels.Start, labels.End, labels.Duration, labels.Description}}
	for _, group := range report.Groups {
		rows = append(rows, []string{group.ProjectName, "", "", "", labels.ProjectTotal, group.TotalDuration, ""})
		for _, row := range group.Rows {
			rows = append(rows, []string{
				row.ProjectName,
				row.TaskName,
				row.Date,
				row.StartTime,
				row.EndTime,
				row.Duration,
				row.Description,
			})
		}
	}
	rows = append(rows, []string{"", "", "", "", labels.Total, report.TotalDuration, ""})

	path, err := reportExportPath("worktime-grouped-by-project", req)
	if err != nil {
		return "", err
	}
	if err := writeSimpleXLSX(path, labels.ReportSheet, rows); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ExportWorktimeTaskDetailsReport(req ReportRequest) (string, error) {
	report, err := a.GetWorktimeTaskDetailsReport(req)
	if err != nil {
		return "", err
	}
	labels := reportExportLabelsFor(req.Language)
	rows := [][]string{{labels.Project, labels.Task, labels.Duration}}
	for _, row := range report.Rows {
		rows = append(rows, []string{row.ProjectName, row.TaskName, row.Duration})
	}
	rows = append(rows, []string{"", labels.Total, report.TotalDuration})

	path, err := reportExportPath("worktime-task-details", req)
	if err != nil {
		return "", err
	}
	if err := writeSimpleXLSX(path, labels.ReportSheet, rows); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ExportWorktimeProjectDetailsReport(req ReportRequest) (string, error) {
	report, err := a.GetWorktimeProjectDetailsReport(req)
	if err != nil {
		return "", err
	}
	labels := reportExportLabelsFor(req.Language)
	rows := [][]string{{labels.Task, labels.Date, labels.Start, labels.End, labels.Duration, labels.Description}}
	for _, row := range report.Rows {
		rows = append(rows, []string{
			row.TaskName,
			row.Date,
			row.StartTime,
			row.EndTime,
			row.Duration,
			row.Description,
		})
	}
	rows = append(rows, []string{"", "", "", labels.Total, report.TotalDuration, ""})

	path, err := reportExportPath("worktime-project-details", req)
	if err != nil {
		return "", err
	}
	if err := writeSimpleXLSX(path, labels.ReportSheet, rows); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ExportTimesheetReport(req ReportRequest) (string, error) {
	report, err := a.GetTimesheetReport(req)
	if err != nil {
		return "", err
	}
	labels := reportExportLabelsFor(req.Language)
	rows := [][]string{}
	if req.Mode == "daily" {
		rows = append(rows, []string{labels.Date, labels.Total, labels.ProjectTime})
		for _, row := range report.DailyRows {
			rows = append(rows, []string{row.Date, row.TotalDuration, row.ProjectDuration})
		}
		rows = append(rows, []string{labels.Total, report.TotalDuration, report.TotalDuration})
	} else {
		rows = append(rows, []string{labels.User, labels.Project, labels.Duration})
		for _, row := range report.ProjectRows {
			rows = append(rows, []string{report.UserName, row.ProjectName, row.Duration})
		}
		rows = append(rows, []string{"", labels.Total, report.TotalDuration})
	}

	path, err := reportExportPath("timesheet", req)
	if err != nil {
		return "", err
	}
	if err := writeSimpleXLSX(path, labels.TimesheetSheet, rows); err != nil {
		return "", err
	}
	return path, nil
}

type reportExportLabels struct {
	Date           string
	Description    string
	Duration       string
	End            string
	Project        string
	ProjectTime    string
	ProjectTotal   string
	ReportSheet    string
	Start          string
	Task           string
	TimesheetSheet string
	Total          string
	User           string
}

func reportExportLabelsFor(language string) reportExportLabels {
	if language == "de" {
		return reportExportLabels{
			Date:           "Datum",
			Description:    "Beschreibung",
			Duration:       "Dauer",
			End:            "Ende",
			Project:        "Projekt",
			ProjectTime:    "Projektzeit",
			ProjectTotal:   "Projektsumme",
			ReportSheet:    "Bericht",
			Start:          "Start",
			Task:           "Aufgabe",
			TimesheetSheet: "Stundenzettel",
			Total:          "Gesamt",
			User:           "Benutzer",
		}
	}
	return reportExportLabels{
		Date:           "Date",
		Description:    "Description",
		Duration:       "Duration",
		End:            "End",
		Project:        "Project",
		ProjectTime:    "Project time",
		ProjectTotal:   "Project total",
		ReportSheet:    "Report",
		Start:          "Start",
		Task:           "Task",
		TimesheetSheet: "Timesheet",
		Total:          "Total",
		User:           "User",
	}
}

func reportExportPath(slug string, req ReportRequest) (string, error) {
	dir, err := paths.DataDir()
	if err != nil {
		return "", err
	}
	exportDir := filepath.Join(dir, "exports")
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(exportDir, fmt.Sprintf("humblebee-%s-%s.xlsx", slug, reportPeriodSlug(req))), nil
}

func reportPeriodSlug(req ReportRequest) string {
	if req.Mode == "daily" {
		return req.StartDate + "_to_" + req.EndDate
	}
	return fmt.Sprintf("%04d-%02d", req.Year, req.Month)
}

func writeSimpleXLSX(path string, sheetName string, rows [][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	files := map[string]string{
		"[Content_Types].xml":        xlsxContentTypes,
		"_rels/.rels":                xlsxRootRels,
		"xl/workbook.xml":            workbookXML(sheetName),
		"xl/_rels/workbook.xml.rels": xlsxWorkbookRels,
		"xl/styles.xml":              xlsxStyles,
		"xl/worksheets/sheet1.xml":   worksheetXML(rows),
		"docProps/core.xml":          xlsxCoreProps,
		"docProps/app.xml":           xlsxAppProps,
	}
	for name, content := range files {
		writer, err := archive.Create(name)
		if err != nil {
			return err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return err
		}
	}
	return archive.Close()
}

func workbookXML(sheetName string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<sheets><sheet name="` + xmlEscape(sheetName) + `" sheetId="1" r:id="rId1"/></sheets></workbook>`
}

func worksheetXML(rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for rowIndex, row := range rows {
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowIndex+1))
		for colIndex, value := range row {
			ref := cellReference(colIndex, rowIndex)
			builder.WriteString(`<c r="` + ref + `" t="inlineStr"><is><t xml:space="preserve">` + xmlEscape(value) + `</t></is></c>`)
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func cellReference(colIndex int, rowIndex int) string {
	col := ""
	for colIndex >= 0 {
		col = string(rune('A'+(colIndex%26))) + col
		colIndex = colIndex/26 - 1
	}
	return fmt.Sprintf("%s%d", col, rowIndex+1)
}

func xmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(value)
}

const xlsxContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
  <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
</Types>`

const xlsxRootRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`

const xlsxWorkbookRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const xlsxStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts>
  <fills count="1"><fill><patternFill patternType="none"/></fill></fills>
  <borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`

const xlsxCoreProps = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <dc:creator>HumbleBee</dc:creator>
</cp:coreProperties>`

const xlsxAppProps = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>HumbleBee</Application>
</Properties>`
