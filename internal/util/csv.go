package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

type VoterCSVRow struct {
	FullName      string
	NIMRaw        string
	NIMNormalized string
	ClassName     string
}

type CSVParseResult struct {
	Rows     []VoterCSVRow
	Rejected []CSVReject
}

type CSVReject struct {
	Row    int
	Reason string
}

// ParseVotersCSV parses a voters CSV file (full_name,nim,class_name).
// Returns valid rows and rejected rows with reasons.
func ParseVotersCSV(r io.Reader) (*CSVParseResult, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	fnIdx, fnOK := colMap["full_name"]
	nimIdx, nimOK := colMap["nim"]
	if !fnOK || !nimOK {
		return nil, fmt.Errorf("CSV must have 'full_name' and 'nim' columns")
	}
	classIdx, classOK := colMap["class_name"]

	result := &CSVParseResult{}
	seen := make(map[string]int) // nim_normalized -> first row
	rowNum := 1                  // 1-based (header is row 0)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Rejected = append(result.Rejected, CSVReject{Row: rowNum + 1, Reason: fmt.Sprintf("parse error: %v", err)})
			rowNum++
			continue
		}
		rowNum++

		fullName := strings.TrimSpace(record[fnIdx])
		if fullName == "" {
			result.Rejected = append(result.Rejected, CSVReject{Row: rowNum, Reason: "full_name is required"})
			continue
		}

		if nimIdx >= len(record) || strings.TrimSpace(record[nimIdx]) == "" {
			result.Rejected = append(result.Rejected, CSVReject{Row: rowNum, Reason: "nim is required"})
			continue
		}

		nimRaw := strings.TrimSpace(record[nimIdx])
		nimNorm := NormalizeNIM(nimRaw)

		if len(nimNorm) > 50 {
			result.Rejected = append(result.Rejected, CSVReject{Row: rowNum, Reason: "nim length exceeds 50 characters"})
			continue
		}

		if firstRow, exists := seen[nimNorm]; exists {
			result.Rejected = append(result.Rejected, CSVReject{Row: rowNum, Reason: fmt.Sprintf("duplicate nim in file (first at row %d)", firstRow)})
			continue
		}
		seen[nimNorm] = rowNum

		var className string
		if classOK && classIdx < len(record) {
			className = strings.TrimSpace(record[classIdx])
		}

		result.Rows = append(result.Rows, VoterCSVRow{
			FullName:      fullName,
			NIMRaw:        nimRaw,
			NIMNormalized: nimNorm,
			ClassName:     className,
		})
	}

	return result, nil
}
