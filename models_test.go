package datagsm

import (
	"encoding/json"
	"testing"
)

func TestStudentNullableFields(t *testing.T) {
	var student Student
	if err := json.Unmarshal([]byte(`{"id":1,"grade":null,"classNum":null,"number":null,"studentNumber":null,"major":null,"specialty":null,"role":"GRADUATE","majorClub":{"id":2,"name":"club","type":"MAJOR_CLUB"}}`), &student); err != nil {
		t.Fatal(err)
	}
	if student.Grade != nil || student.ClassNum != nil || student.Number != nil || student.StudentNumber != nil || student.Major != nil || student.Specialty != nil || student.MajorClub.ID != 2 {
		t.Fatal("nullable fields or club summary were lost")
	}
}
