package main

import (
	"context"
	"fmt"
	"log"
	"os"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
	"github.com/tmp-it-play/datagsm-sdk-go/openapi"
)

func main() {
	client, err := openapi.NewClient(openapi.Config{APIKey: os.Getenv("DATAGSM_API_KEY")})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	page, err := client.Students.List(context.Background(), openapi.StudentQuery{Grade: datagsm.Ptr(1), Size: 20})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("학생 수: %d\n", page.TotalElements)
	for _, student := range page.Students {
		fmt.Println(student.ID, student.Name)
	}
}
