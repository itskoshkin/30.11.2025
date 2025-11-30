package pdf

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

func GeneratePDF(sets []int, text [][]string) (string, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Link Availability Report")
	pdf.Ln(10)

	for i, num := range sets {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, fmt.Sprintf("Set #%d", num))
		pdf.Ln(10)

		pdf.SetFont("Courier", "", 14)
		for _, line := range text[i] {
			pdf.Cell(40, 10, line)
			pdf.Ln(10)
		}
	}

	var filePath string
	if len(sets) == 1 {
		filePath = fmt.Sprintf("files/set_%d.pdf", sets[0])
	} else {
		strSets := make([]string, len(sets))
		for i, n := range sets {
			strSets[i] = strconv.Itoa(n)
		}
		filePath = fmt.Sprintf("files/sets_%s.pdf", strings.Join(strSets, "-"))
	}

	return filePath, pdf.OutputFileAndClose(filePath)
}
