// Package event verifies and decodes DataGSM webhook events.
package event

import datagsm "github.com/tmp-it-play/datagsm-sdk-go"

type Type string

const (
	StudentUpdated Type = "student.updated"
	ClubUpdated    Type = "club.updated"
	ProjectUpdated Type = "project.updated"
)

type Student struct {
	ID                 int64               `json:"student_id"`
	Name               string              `json:"name"`
	Email              string              `json:"email"`
	Sex                datagsm.Sex         `json:"sex"`
	Grade              *int                `json:"grade"`
	ClassNum           *int                `json:"class_num"`
	Number             *int                `json:"number"`
	StudentNumber      *int                `json:"student_number"`
	Major              *datagsm.Major      `json:"major"`
	Specialty          *string             `json:"specialty"`
	Role               datagsm.StudentRole `json:"role"`
	DormitoryFloor     *int                `json:"dormitory_floor"`
	DormitoryRoom      *int                `json:"dormitory_room"`
	MajorClubName      *string             `json:"major_club_name"`
	AutonomousClubName *string             `json:"autonomous_club_name"`
	GitHubID           *string             `json:"github_id"`
}

type Participant struct {
	StudentNumber *int   `json:"student_number"`
	Name          string `json:"name"`
}

type Club struct {
	ID            int64              `json:"club_id"`
	Name          string             `json:"name"`
	Type          datagsm.ClubType   `json:"type"`
	FoundedYear   int                `json:"founded_year"`
	Status        datagsm.ClubStatus `json:"status"`
	AbolishedYear *int               `json:"abolished_year"`
	Leader        *Participant       `json:"leader"`
	Participants  []Participant      `json:"participants"`
}

type ClubSummary struct {
	ID   int64  `json:"club_id"`
	Name string `json:"name"`
}

type Project struct {
	ID            int64                     `json:"project_id"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	StartYear     int                       `json:"start_year"`
	EndYear       *int                      `json:"end_year"`
	Status        datagsm.ProjectStatus     `json:"status"`
	Categories    []datagsm.ProjectCategory `json:"categories"`
	DeploymentURL *string                   `json:"deployment_url"`
	Club          *ClubSummary              `json:"club"`
	Participants  []Participant             `json:"participants"`
	Repositories  []string                  `json:"repositories"`
	TechStacks    []string                  `json:"tech_stacks"`
}
