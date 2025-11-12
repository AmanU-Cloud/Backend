package file

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Caritas-Team/reviewer/internal/model"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// PDFParser парсит PDF файлы и извлекает данные согласно OpenAPI схеме
type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// ParsePDF парсит PDF файл и возвращает ChildProfile
func (p *PDFParser) ParsePDF(reader io.Reader) (*model.ChildProfile, error) {
	// Читаем весь файл в память
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	// Создаем временный файл для pdfcpu
	tmpFile, err := os.CreateTemp("", "pdf_parse_*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Извлекаем текст используя pdfcpu extract команду
	text, err := extractTextWithPDFCPU(tmpFile.Name())
	if err != nil {
		slog.Warn("Failed to extract text with pdfcpu, trying alternative method", "err", err)
		// Альтернативный метод: используем ReadContext для валидации и базового парсинга
		ctx, err := api.ReadContext(bytes.NewReader(data), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF context: %w", err)
		}
		slog.Info("PDF validated", "pages", ctx.PageCount)
		// Если не удалось извлечь текст, возвращаем пустой профиль
		text = ""
	}

	// Парсим структурированные данные из текста
	profile := &model.ChildProfile{
		Portrait:               p.parsePortrait(text),
		BasicAssessment:        p.parseBasicAssessment(text),
		LanguageAssessment:     p.parseLanguageAssessment(text),
		CommunicationFunctions: p.parseCommunicationFunctions(text),
		AACUsage:               p.parseAACUsage(text),
		Vocabulary:             p.parseVocabulary(text),
		InterestsAndBarriers:   p.parseInterestsAndBarriers(text),
		CommunicationCircles:   p.parseCommunicationCircles(text),
	}

	return profile, nil
}

// extractTextWithPDFCPU извлекает текст из PDF используя pdfcpu CLI
func extractTextWithPDFCPU(pdfPath string) (string, error) {
	// Создаем временную директорию для извлечения текста
	tmpDir, err := os.MkdirTemp("", "pdf_extract_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Используем pdfcpu extract для извлечения текста
	cmd := exec.Command("pdfcpu", "extract", "-mode", "text", pdfPath, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pdfcpu extract failed: %w, output: %s", err, string(output))
	}

	// Читаем извлеченный текст из файлов в tmpDir
	var fullText strings.Builder
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".txt") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fullText.Write(data)
			fullText.WriteString("\n")
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to read extracted text: %w", err)
	}

	return fullText.String(), nil
}

// parsePortrait извлекает данные портрета пользователя
func (p *PDFParser) parsePortrait(text string) *model.Portrait {
	portrait := &model.Portrait{}

	// Ищем имя ребенка
	if match := findField(text, "Имя обследуемого", "Имя", "Ребенок"); match != "" {
		portrait.ChildName = match
	}

	// Ищем ФИ родителя
	if match := findField(text, "ФИ родителя", "Родитель", "Опекун"); match != "" {
		portrait.ParentName = match
	}

	// Ищем дату заполнения
	if match := findDateField(text, "Дата заполнения", "Дата"); match != "" {
		portrait.DateFilled = match
	}

	// Ищем дату рождения
	if match := findDateField(text, "Дата рождения", "Родился"); match != "" {
		portrait.DateOfBirth = match
	}

	// Ищем диагноз
	if match := findField(text, "Диагноз", "Диагноз:"); match != "" {
		portrait.Diagnosis = match
	}

	// Ищем социальную ситуацию
	if match := findField(text, "Социальная ситуация", "Особенности социальной ситуации"); match != "" {
		portrait.SocialSituation = match
	}

	// Ищем место проживания
	if match := findField(text, "Место проживания", "Где проживает"); match != "" {
		portrait.PlaceOfResidence = match
	}

	return portrait
}

// parseBasicAssessment извлекает базовую оценку
func (p *PDFParser) parseBasicAssessment(text string) *model.BasicAssessment {
	assessment := &model.BasicAssessment{}

	// Парсим вербальную речь
	assessment.VerbalSpeech = p.parseVerbalSpeech(text)

	// Парсим письменную речь
	assessment.WrittenSpeech = p.parseWrittenSpeech(text)

	// Парсим зрение
	assessment.Vision = p.parseVision(text)

	// Парсим слух
	assessment.Hearing = p.parseHearing(text)

	// Парсим понимание речи
	assessment.UnderstandingSpeech = p.parseUnderstandingSpeech(text)

	// Парсим моторные навыки
	assessment.MotorSkills = p.parseMotorSkills(text)

	return assessment
}

func (p *PDFParser) parseVerbalSpeech(text string) *model.VerbalSpeech {
	speech := &model.VerbalSpeech{}

	if match := findField(text, "Вербальная речь", "Особенности вербальной речи"); match != "" {
		speech.MainFeature = match
	}

	// Ищем дополнительные заметки
	notes := findListItems(text, "Дополнительно", "Заметки")
	speech.AdditionalNotes = notes

	return speech
}

func (p *PDFParser) parseWrittenSpeech(text string) *model.WrittenSpeech {
	speech := &model.WrittenSpeech{}

	if match := findField(text, "Письменная речь", "Статус письменной речи"); match != "" {
		speech.Status = match
	}

	return speech
}

func (p *PDFParser) parseVision(text string) *model.Vision {
	vision := &model.Vision{}

	if match := findField(text, "Зрение", "Статус зрения"); match != "" {
		vision.GeneralStatus = match
	}

	vision.SpecificIssues = findListItems(text, "Проблемы со зрением", "Проблемы")
	vision.Capabilities = findListItems(text, "Возможности зрения", "Видит")

	return vision
}

func (p *PDFParser) parseHearing(text string) *model.Hearing {
	hearing := &model.Hearing{}

	if match := findField(text, "Слух", "Статус слуха"); match != "" {
		hearing.GeneralStatus = match
	}

	hearing.AdditionalNotes = findListItems(text, "Дополнительно", "Заметки")

	return hearing
}

func (p *PDFParser) parseUnderstandingSpeech(text string) *model.UnderstandingSpeech {
	understanding := &model.UnderstandingSpeech{}

	understanding.Capabilities = findListItems(text, "Понимание речи", "Понимает")

	return understanding
}

func (p *PDFParser) parseMotorSkills(text string) *model.MotorSkills {
	skills := &model.MotorSkills{}

	if match := findField(text, "Моторные навыки", "Моторная сфера"); match != "" {
		skills.Description = match
	}

	skills.PointingMethods = findListItems(text, "Способы указания", "Указывает")

	return skills
}

// parseLanguageAssessment извлекает языковую оценку
func (p *PDFParser) parseLanguageAssessment(text string) *model.LanguageAssessment {
	assessment := &model.LanguageAssessment{}

	// Парсим уровни применения языковых навыков
	assessment.LanguageLevelApplication = p.parseLanguageLevelApplication(text)

	// Парсим инициативу
	assessment.Initiative = p.parseInitiative(text)

	// Парсим сводку по коммуникативным функциям
	assessment.CommunicationFunctionsSummary = p.parseCommunicationFunctionsSummary(text)

	return assessment
}

func (p *PDFParser) parseLanguageLevelApplication(text string) *model.LanguageLevelApplication {
	levels := &model.LanguageLevelApplication{}

	levels.DointencionalCommunication = findPercent(text, "Доинтенциональная коммуникация", "Доинтенциональная")
	levels.Protolanguage = findPercent(text, "Протоязык", "Протоязык")
	levels.Holoprasis = findPercent(text, "Голофраза", "Голофраза")
	levels.Phrase = findPercent(text, "Фраза", "Фраза")

	return levels
}

func (p *PDFParser) parseInitiative(text string) *model.Initiative {
	initiative := &model.Initiative{}

	initiative.Level1 = findPercent(text, "Уровень 1", "Инициатива уровень 1")
	initiative.Level2 = findPercent(text, "Уровень 2", "Инициатива уровень 2")
	initiative.Level3 = findPercent(text, "Уровень 3", "Инициатива уровень 3")

	return initiative
}

func (p *PDFParser) parseCommunicationFunctionsSummary(text string) *model.CommunicationFunctionsSummary {
	summary := &model.CommunicationFunctionsSummary{}

	summary.RefusalRejection = findPercent(text, "Отказ/отклонение", "Отказывается")
	summary.ObtainingDesired = findPercent(text, "Получение желаемого", "Получает")
	summary.SocialInteraction = findPercent(text, "Социальное взаимодействие", "Социальное")
	summary.InformationExchange = findPercent(text, "Обмен информацией", "Обмен")

	return summary
}

// parseCommunicationFunctions извлекает коммуникативные функции
func (p *PDFParser) parseCommunicationFunctions(text string) []model.CommunicationFunction {
	functions := []model.CommunicationFunction{}

	// Ищем основные функции
	functionNames := []string{
		"ОТКАЗЫВАЕТСЯ, ОТКЛОНЯЕТ",
		"ПОЛУЧЕНИЕ ЖЕЛАЕМОГО",
		"СОЦИАЛЬНОЕ ВЗАИМОДЕЙСТВИЕ",
		"ОБМЕН ИНФОРМАЦИЕЙ",
	}

	for _, name := range functionNames {
		if strings.Contains(text, name) {
			fn := model.CommunicationFunction{
				FunctionName: name,
				Table:        p.parseCommunicationFunctionTable(text, name),
			}
			functions = append(functions, fn)
		}
	}

	return functions
}

func (p *PDFParser) parseCommunicationFunctionTable(text string, functionName string) []model.CommunicationFunctionEntry {
	// Упрощенный парсинг таблицы
	// В реальной реализации нужно более точно извлекать данные из таблицы
	entries := []model.CommunicationFunctionEntry{}

	// Ищем строки таблицы для данной функции
	// Это упрощенная версия, в реальности нужен более сложный парсинг

	return entries
}

// parseAACUsage извлекает данные об использовании АДК
func (p *PDFParser) parseAACUsage(text string) *model.AACUsage {
	aac := &model.AACUsage{}

	aac.AACTools = findListItems(text, "Средства АДК", "Инструменты")

	if match := findField(text, "Степень поддержки", "Поддержка"); match != "" {
		aac.SupportLevel = match
	}

	if match := findField(text, "Способ доступа", "Доступ"); match != "" {
		aac.AccessMethod = match
	}

	if match := findField(text, "Смартфон/планшет", "Взаимодействие"); match != "" {
		aac.SmartphoneTabletInteraction = match
	}

	return aac
}

// parseVocabulary извлекает словарный запас
func (p *PDFParser) parseVocabulary(text string) *model.Vocabulary {
	vocab := &model.Vocabulary{}

	vocab.BaseVocabulary = findListItems(text, "Базовый словарь", "Слова")
	vocab.CustomWords = findListItems(text, "Собственные слова", "Особые слова")

	// Подсчитываем общее количество слов
	vocab.TotalWords = len(vocab.BaseVocabulary) + len(vocab.CustomWords)

	return vocab
}

// parseInterestsAndBarriers извлекает интересы и барьеры
func (p *PDFParser) parseInterestsAndBarriers(text string) *model.InterestsAndBarriers {
	ib := &model.InterestsAndBarriers{}

	ib.PreferredActivities = findListItems(text, "Предпочитаемые активности", "Нравится")
	ib.UncomfortableActivities = findListItems(text, "Неудобные активности", "Не нравится")

	return ib
}

// parseCommunicationCircles извлекает круги общения
func (p *PDFParser) parseCommunicationCircles(text string) []model.CommunicationCircle {
	circles := []model.CommunicationCircle{}

	contactTypes := []string{"Мама", "Папа", "Брат/сестра", "Друзья", "Знакомые", "Специалисты", "Другое"}

	for _, contactType := range contactTypes {
		if strings.Contains(text, contactType) {
			circle := model.CommunicationCircle{
				ContactType: contactType,
			}

			if match := findField(text, contactType+" стиль", "Стиль"); match != "" {
				circle.CommunicationStyle = match
			}

			if match := findField(text, contactType+" реакция", "Реакция"); match != "" {
				circle.SignalResponse = match
			}

			if match := findField(text, contactType+" диалог", "Диалог"); match != "" {
				circle.DialogueSupport = match
			}

			circles = append(circles, circle)
		}
	}

	return circles
}

// Вспомогательные функции для парсинга

func findField(text string, patterns ...string) string {
	for _, pattern := range patterns {
		re := regexp.MustCompile(fmt.Sprintf(`(?i)%s[:\s]+([^\n]+)`, regexp.QuoteMeta(pattern)))
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func findDateField(text string, patterns ...string) string {
	for _, pattern := range patterns {
		// Ищем дату в формате ДД.ММ.ГГГГ или ДД-ММ-ГГГГ
		re := regexp.MustCompile(fmt.Sprintf(`(?i)%s[:\s]+(\d{1,2}[.\-]\d{1,2}[.\-]\d{4})`, regexp.QuoteMeta(pattern)))
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func findPercent(text string, patterns ...string) int {
	for _, pattern := range patterns {
		// Ищем процент после паттерна
		re := regexp.MustCompile(fmt.Sprintf(`(?i)%s[:\s]+(\d+)\s*%%?`, regexp.QuoteMeta(pattern)))
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				return val
			}
		}
	}
	return 0
}

func findListItems(text string, patterns ...string) []string {
	items := []string{}
	for _, pattern := range patterns {
		// Ищем список после паттерна
		re := regexp.MustCompile(fmt.Sprintf(`(?i)%s[:\s]+(.*?)(?:\n\n|\n[A-ZА-Я]|$)`, regexp.QuoteMeta(pattern)))
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			// Разбиваем на элементы списка
			parts := strings.Split(matches[1], "\n")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" && (strings.HasPrefix(part, "-") || strings.HasPrefix(part, "•") || strings.HasPrefix(part, "*")) {
					part = strings.TrimLeft(part, "-•* ")
					if part != "" {
						items = append(items, part)
					}
				}
			}
		}
	}
	return items
}
