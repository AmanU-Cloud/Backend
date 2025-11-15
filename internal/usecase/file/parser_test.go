package file

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPdf(t *testing.T) {
	file, err := os.Open("../../../docs/example.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	parser := PDFParser{}
	ch, err := parser.ParsePDF(*file)
	assert.NoError(t, err)
	assert.NotNil(t, ch)
}
