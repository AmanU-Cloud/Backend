package model

import "time"

// Структуры данных согласно OpenAPI

type UploadResponse struct {
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`
	PairsCount  int    `json:"pairs_count"`
	Message     string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type DuplicateOperationError struct {
	Error               string `json:"error"`
	Code                string `json:"code"`
	ExistingOperationID string `json:"existing_operation_id"`
}

type Progress struct {
	Processed int `json:"processed"`
	Total     int `json:"total"`
}

type OperationInProgress struct {
	OperationID                string    `json:"operation_id"`
	Status                     string    `json:"status"`
	Progress                   *Progress `json:"progress,omitempty"`
	EstimatedCompletionSeconds int       `json:"estimated_completion_seconds,omitempty"`
}

type OperationComplete struct {
	OperationID string            `json:"operation_id"`
	Status      string            `json:"status"`
	Results     []ChildComparison `json:"results"`
}

type OperationError struct {
	OperationID string   `json:"operation_id"`
	Status      string   `json:"status"`
	Error       string   `json:"error"`
	ErrorCode   string   `json:"error_code,omitempty"`
	FailedFiles []string `json:"failed_files,omitempty"`
}

// Структуры для данных из PDF (согласно OpenAPI)

type ChildComparison struct {
	ChildID string          `json:"child_id"`
	Before  ChildProfile    `json:"before"`
	After   ChildProfile    `json:"after"`
	Changes ChangesAnalysis `json:"changes,omitempty"`
}

type ChildProfile struct {
	Portrait               *Portrait               `json:"portrait,omitempty"`
	BasicAssessment        *BasicAssessment        `json:"basic_assessment,omitempty"`
	LanguageAssessment     *LanguageAssessment     `json:"language_assessment,omitempty"`
	CommunicationFunctions []CommunicationFunction `json:"communication_functions,omitempty"`
	AACUsage               *AACUsage               `json:"aac_usage,omitempty"`
	Vocabulary             *Vocabulary             `json:"vocabulary,omitempty"`
	InterestsAndBarriers   *InterestsAndBarriers   `json:"interests_and_barriers,omitempty"`
	CommunicationCircles   []CommunicationCircle   `json:"communication_circles,omitempty"`
}

type Portrait struct {
	ChildName        string `json:"child_name"`
	ParentName       string `json:"parent_name"`
	DateFilled       string `json:"date_filled,omitempty"`
	DateOfBirth      string `json:"date_of_birth,omitempty"`
	Diagnosis        string `json:"diagnosis,omitempty"`
	SocialSituation  string `json:"social_situation,omitempty"`
	PlaceOfResidence string `json:"place_of_residence,omitempty"`
}

type BasicAssessment struct {
	VerbalSpeech        *VerbalSpeech        `json:"verbal_speech,omitempty"`
	WrittenSpeech       *WrittenSpeech       `json:"written_speech,omitempty"`
	Vision              *Vision              `json:"vision,omitempty"`
	Hearing             *Hearing             `json:"hearing,omitempty"`
	UnderstandingSpeech *UnderstandingSpeech `json:"understanding_speech,omitempty"`
	MotorSkills         *MotorSkills         `json:"motor_skills,omitempty"`
}

type VerbalSpeech struct {
	MainFeature     string   `json:"main_feature"`
	AdditionalNotes []string `json:"additional_notes,omitempty"`
}

type WrittenSpeech struct {
	Status string `json:"status"`
}

type Vision struct {
	GeneralStatus  string   `json:"general_status"`
	SpecificIssues []string `json:"specific_issues,omitempty"`
	Capabilities   []string `json:"capabilities,omitempty"`
}

type Hearing struct {
	GeneralStatus   string   `json:"general_status"`
	AdditionalNotes []string `json:"additional_notes,omitempty"`
}

type UnderstandingSpeech struct {
	Capabilities []string `json:"capabilities"`
}

type MotorSkills struct {
	Description     string   `json:"description"`
	PointingMethods []string `json:"pointing_methods,omitempty"`
}

type LanguageAssessment struct {
	LanguageLevelApplication      *LanguageLevelApplication      `json:"language_level_application,omitempty"`
	Initiative                    *Initiative                    `json:"initiative,omitempty"`
	CommunicationFunctionsSummary *CommunicationFunctionsSummary `json:"communication_functions_summary,omitempty"`
}

type LanguageLevelApplication struct {
	DointencionalCommunication int `json:"dointencional_communication"`
	Protolanguage              int `json:"protolanguage"`
	Holoprasis                 int `json:"holoprasis"`
	Phrase                     int `json:"phrase"`
}

type Initiative struct {
	Level1 int `json:"level_1"`
	Level2 int `json:"level_2"`
	Level3 int `json:"level_3"`
}

type CommunicationFunctionsSummary struct {
	RefusalRejection    int `json:"refusal_rejection"`
	ObtainingDesired    int `json:"obtaining_desired"`
	SocialInteraction   int `json:"social_interaction"`
	InformationExchange int `json:"information_exchange"`
}

type CommunicationFunction struct {
	FunctionName string                       `json:"function_name"`
	Table        []CommunicationFunctionEntry `json:"table,omitempty"`
}

type CommunicationFunctionEntry struct {
	Level             string `json:"level"`
	FormedPercent     int    `json:"formed_percent"`
	InitiativePercent int    `json:"initiative_percent"`
	FrequencyCategory string `json:"frequency_category"`
	FrequencyPercent  int    `json:"frequency_percent"`
}

type AACUsage struct {
	AACTools                    []string `json:"aac_tools,omitempty"`
	SupportLevel                string   `json:"support_level,omitempty"`
	AccessMethod                string   `json:"access_method,omitempty"`
	SmartphoneTabletInteraction string   `json:"smartphone_tablet_interaction,omitempty"`
}

type Vocabulary struct {
	BaseVocabulary []string `json:"base_vocabulary,omitempty"`
	CustomWords    []string `json:"custom_words,omitempty"`
	TotalWords     int      `json:"total_words,omitempty"`
}

type InterestsAndBarriers struct {
	PreferredActivities     []string `json:"preferred_activities,omitempty"`
	UncomfortableActivities []string `json:"uncomfortable_activities,omitempty"`
}

type CommunicationCircle struct {
	ContactType        string `json:"contact_type"`
	CommunicationStyle string `json:"communication_style,omitempty"`
	SignalResponse     string `json:"signal_response,omitempty"`
	DialogueSupport    string `json:"dialogue_support,omitempty"`
}

type ChangesAnalysis struct {
	LanguageImprovements         *LanguageImprovements `json:"language_improvements,omitempty"`
	CommunicationFunctionChanges map[string]int        `json:"communication_function_changes,omitempty"`
	NewVocabularyWords           []string              `json:"new_vocabulary_words,omitempty"`
	LostVocabularyWords          []string              `json:"lost_vocabulary_words,omitempty"`
	ImprovedSkills               []string              `json:"improved_skills,omitempty"`
	AreasForImprovement          []string              `json:"areas_for_improvement,omitempty"`
}

type LanguageImprovements struct {
	DointencionalCommunicationChange int `json:"dointencional_communication_change"`
	ProtolanguageChange              int `json:"protolanguage_change"`
	HoloprasisChange                 int `json:"holoprasis_change"`
	PhraseChange                     int `json:"phrase_change"`
}

// OperationStatus хранит статус операции
type OperationStatus struct {
	OperationID string            `json:"operation_id"`
	Status      string            `json:"status"` // NEW, PROGRESS, DONE, ERROR
	Progress    *Progress         `json:"progress,omitempty"`
	Results     []ChildComparison `json:"results,omitempty"`
	Error       string            `json:"error,omitempty"`
	ErrorCode   string            `json:"error_code,omitempty"`
	FailedFiles []string          `json:"failed_files,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	Files       []string          `json:"files,omitempty"` // Пути к файлам
}
