package main

import (
	"os"
	"fmt"
	"flag"
	
	"github.com/pkg/errors"
	//"github.com/metalnem/runtastic/api"
)

var (
	email	 = flag.String("email","","")
	password = flag.String("password","","")
	
	errMissingCredentials = errors.New("Missing email address or password")
)

const usage = `Usage of Runtastic->Strava Importer:
  -email string
		Email (required)
		
  -password string
		Password(required)`
		
func getCredentials() (string, string, error) {
	email := *email
	password := *password
	
	if email != "" && password != "" {
		return email, password, nil
	}
	
	email = os.Getenv("RUNTASTIC_EMAIL")
	password = os.Getenv("RUNTASTIC_PASSWORD")
	
	if email != "" && password != "" {
		return email, password, nil
	}
	
	return "", "", errMissingCredentials
}

func main() {
	flag.Parse()
	
	email, password, err := getCredentials()
	
	if err != nil {
		fmt.Println(usage)
		os.Exit(1)
	}
	
	fmt.Println(email)
	fmt.Println(password)
	
}