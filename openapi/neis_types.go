package openapi

type MealType string

const (
	MealTypeBreakfast MealType = "BREAKFAST"
	MealTypeLunch     MealType = "LUNCH"
	MealTypeDinner    MealType = "DINNER"
)

type Meal struct {
	MealID          string   `json:"mealId"`
	SchoolCode      string   `json:"schoolCode"`
	SchoolName      string   `json:"schoolName"`
	OfficeCode      string   `json:"officeCode"`
	OfficeName      string   `json:"officeName"`
	MealDate        string   `json:"mealDate"`
	MealType        MealType `json:"mealType"`
	MealMenu        []string `json:"mealMenu"`
	MealAllergyInfo []string `json:"mealAllergyInfo,omitempty"`
	MealCalories    *string  `json:"mealCalories,omitempty"`
	OriginInfo      *string  `json:"originInfo,omitempty"`
	NutritionInfo   *string  `json:"nutritionInfo,omitempty"`
	MealServeCount  *int     `json:"mealServeCount,omitempty"`
}

type Schedule struct {
	ScheduleID       string  `json:"scheduleId"`
	SchoolCode       string  `json:"schoolCode"`
	SchoolName       string  `json:"schoolName"`
	OfficeCode       string  `json:"officeCode"`
	OfficeName       string  `json:"officeName"`
	ScheduleDate     string  `json:"scheduleDate"`
	AcademicYear     string  `json:"academicYear"`
	EventName        string  `json:"eventName"`
	EventContent     *string `json:"eventContent,omitempty"`
	DayCategory      *string `json:"dayCategory,omitempty"`
	SchoolCourseType *string `json:"schoolCourseType,omitempty"`
	DayNightType     *string `json:"dayNightType,omitempty"`
	TargetGrades     []int   `json:"targetGrades"`
}

type Timetable struct {
	TimetableID   string  `json:"timetableId"`
	SchoolCode    string  `json:"schoolCode"`
	SchoolName    string  `json:"schoolName"`
	OfficeCode    string  `json:"officeCode"`
	OfficeName    string  `json:"officeName"`
	TimetableDate string  `json:"timetableDate"`
	AcademicYear  string  `json:"academicYear"`
	Semester      *string `json:"semester"`
	Grade         int     `json:"grade"`
	ClassNum      int     `json:"classNum"`
	Period        int     `json:"period"`
	Subject       *string `json:"subject,omitempty"`
}
