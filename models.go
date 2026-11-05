package datagsm

type Major string

const (
	MajorSWDevelopment Major = "SW_DEVELOPMENT"
	MajorSmartIoT      Major = "SMART_IOT"
	MajorAI            Major = "AI"
)

type Sex string

const (
	SexMan   Sex = "MAN"
	SexWoman Sex = "WOMAN"
)

type SortDirection string

const (
	SortDirASC  SortDirection = "ASC"
	SortDirDESC SortDirection = "DESC"
)

type ClubType string

const (
	ClubTypeMajor      ClubType = "MAJOR_CLUB"
	ClubTypeAutonomous ClubType = "AUTONOMOUS_CLUB"
)

type ClubStatus string

const (
	ClubStatusActive    ClubStatus = "ACTIVE"
	ClubStatusAbolished ClubStatus = "ABOLISHED"
)

type ClubSortBy string

const (
	ClubSortByID          ClubSortBy = "ID"
	ClubSortByName        ClubSortBy = "NAME"
	ClubSortByType        ClubSortBy = "TYPE"
	ClubSortByFoundedYear ClubSortBy = "FOUNDED_YEAR"
	ClubSortByStatus      ClubSortBy = "STATUS"
)

type ClubSummary struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Type ClubType `json:"type"`
}

type ParticipantInfo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	StudentNumber *int   `json:"studentNumber"`
	Major         *Major `json:"major"`
	Sex           Sex    `json:"sex"`
}

type Club struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	Type          ClubType          `json:"type"`
	Status        ClubStatus        `json:"status"`
	FoundedYear   int               `json:"foundedYear"`
	AbolishedYear *int              `json:"abolishedYear"`
	Leader        *ParticipantInfo  `json:"leader"`
	Participants  []ParticipantInfo `json:"participants"`
}

type StudentRole string

const (
	StudentRoleGeneralStudent   StudentRole = "GENERAL_STUDENT"
	StudentRoleStudentCouncil   StudentRole = "STUDENT_COUNCIL"
	StudentRoleDormitoryManager StudentRole = "DORMITORY_MANAGER"
	StudentRoleGraduate         StudentRole = "GRADUATE"
	StudentRoleWithdrawn        StudentRole = "WITHDRAWN"
)

type StudentSortBy string

const (
	StudentSortByID            StudentSortBy = "ID"
	StudentSortByName          StudentSortBy = "NAME"
	StudentSortByEmail         StudentSortBy = "EMAIL"
	StudentSortByStudentNumber StudentSortBy = "STUDENT_NUMBER"
	StudentSortByGrade         StudentSortBy = "GRADE"
	StudentSortByClassNum      StudentSortBy = "CLASS_NUM"
	StudentSortByNumber        StudentSortBy = "NUMBER"
	StudentSortByMajor         StudentSortBy = "MAJOR"
	StudentSortByRole          StudentSortBy = "ROLE"
	StudentSortBySex           StudentSortBy = "SEX"
	StudentSortByDormitoryRoom StudentSortBy = "DORMITORY_ROOM"
)

type Student struct {
	ID             int64        `json:"id"`
	Name           string       `json:"name"`
	Sex            Sex          `json:"sex"`
	Email          string       `json:"email"`
	Grade          *int         `json:"grade"`
	ClassNum       *int         `json:"classNum"`
	Number         *int         `json:"number"`
	StudentNumber  *int         `json:"studentNumber"`
	Major          *Major       `json:"major"`
	Specialty      *string      `json:"specialty"`
	Role           StudentRole  `json:"role"`
	DormitoryFloor *int         `json:"dormitoryFloor"`
	DormitoryRoom  *int         `json:"dormitoryRoom"`
	MajorClub      *ClubSummary `json:"majorClub"`
	AutonomousClub *ClubSummary `json:"autonomousClub"`
	GitHubID       *string      `json:"githubId"`
	GitHubURL      *string      `json:"githubUrl"`
}

type ProjectSortBy string

const (
	ProjectSortByID   ProjectSortBy = "ID"
	ProjectSortByName ProjectSortBy = "NAME"
)

type Project struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	StartYear     int               `json:"startYear"`
	EndYear       *int              `json:"endYear"`
	Status        ProjectStatus     `json:"status"`
	Categories    []ProjectCategory `json:"categories"`
	IconURL       *string           `json:"iconUrl"`
	IconKey       *string           `json:"iconKey"`
	DeploymentURL *string           `json:"deploymentUrl"`
	Repositories  []string          `json:"repositories"`
	TechStacks    []string          `json:"techStacks"`
	Club          *ClubSummary      `json:"club"`
	Participants  []ParticipantInfo `json:"participants"`
}

type ProjectStatus string

const (
	ProjectStatusActive ProjectStatus = "ACTIVE"
	ProjectStatusEnded  ProjectStatus = "ENDED"
)

type ProjectCategory string

const (
	ProjectCategoryToy          ProjectCategory = "TOY_PROJECT"
	ProjectCategoryTeam         ProjectCategory = "TEAM"
	ProjectCategoryClub         ProjectCategory = "CLUB"
	ProjectCategoryIdeaFestival ProjectCategory = "IDEA_FESTIVAL"
	ProjectCategoryNotDeployed  ProjectCategory = "NOT_DEPLOYED"
)
