package file

import (
	"os"

	"github.com/Caritas-Team/reviewer/internal/model"
)

// ParsePairPDFs парсит два PDF-файла и возвращает профили до/после
func ParsePairPDFs(parser *PDFParser, before, after os.File) (*model.ChildProfile, *model.ChildProfile, error) {
	profileBefore, err := parser.ParsePDF(before)
	if err != nil {
		return nil, nil, err
	}
	profileAfter, err := parser.ParsePDF(after)
	if err != nil {
		return profileBefore, nil, err
	}
	return profileBefore, profileAfter, nil
}
