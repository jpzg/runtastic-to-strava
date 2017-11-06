package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jpzg/runtastic-to-strava/upload"
	"github.com/metalnem/runtastic/api"
	"github.com/pkg/errors"
	"github.com/strava/go.strava"
)

var (
	email    = flag.String("email", "", "Email address for your Runtastic account. May be stored in the environment variable RUNTASTIC_EMAIL.")
	password = flag.String("password", "", "Password for your Runtastic account. May be stored in the environment variable RUNTASTIC_PASSWORD.")
	token    = flag.String("token", "", "The Strava access token for your account. See strava.com/settings/api. May be stored in the environment variable STRAVA_TOKEN.")
	thorough = flag.Bool("thorough", false, "Check that all Runtastic activities are on Strava (not just all activities until the last recorded Strava activity)")

	errMissingCredentials = errors.New("Missing email address, password, or access token")
)

const usage = `Usage of Runtastic->Strava Importer:
  -email string
		Email (required)
		
  -password string
		Password(required)
		
  -token string
		Strava access token for your account (required)
		
  -thorough
		Checks that all activities on the Runtastic account are also on Strava. Normal check goes only until the last recorded Strava activity`

func getInput() (string, string, string, bool, error) {
	email := *email
	password := *password
	token := *token
	thorough := *thorough

	if email != "" && password != "" && token != "" {
		return email, password, token, thorough, nil
	}

	email = os.Getenv("RUNTASTIC_EMAIL")
	password = os.Getenv("RUNTASTIC_PASSWORD")
	token = os.Getenv("STRAVA_TOKEN")

	if email != "" && password != "" && token != "" {
		return email, password, token, thorough, nil
	}

	return "", "", "", false, errMissingCredentials
}

func main() {
	flag.Parse()

	email, password, token, thorough, err := getInput()

	if err != nil {
		fmt.Println(usage)
		os.Exit(1)
	}

	if thorough {
		fmt.Println("Thorough mode is on")
	}

	ctx := context.Background()
	session, err := api.Login(ctx, email, password)

	if err != nil {
		glog.Exit(err)
	}

	fmt.Println("Successfully logged into Runtastic")

	client := strava.NewClient(token)
	athlete := strava.NewCurrentAthleteService(client)

	switch thorough {
	case true:
		count, err := upload.UploadThorough(session, ctx, athlete)
	case false:
		count, err := upload.UploadNormal(session, ctx, athlete)
	}

	fmt.Printf("\nUploaded %d activities\n", count)

	if err != nil {
		glog.Exit(err)
	}

	glog.Flush()

}
