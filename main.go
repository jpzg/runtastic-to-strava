package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jpzg/runtastic-to-strava/oauth"
	"github.com/jpzg/runtastic-to-strava/upload"
	"github.com/metalnem/runtastic/api"
	"github.com/pkg/errors"
	"github.com/strava/go.strava"
)

var (
	email    = flag.String("email", "", "Email address for your Runtastic account. May be stored in the environment variable RUNTASTIC_EMAIL.")
	password = flag.String("password", "", "Password for your Runtastic account. May be stored in the environment variable RUNTASTIC_PASSWORD.")
	token    = flag.String("token", "", "The Strava access token for your account. See strava.com/settings/api. May be stored in the environment variable STRAVA_TOKEN.")
	id       = flag.String("id", "", "The client ID associated with your API application. See strava.com/settings/api.")
	secret   = flag.String("secret", "", "The client secret associated with your API application. See strava.com/settings/api.")

	thorough = flag.Bool("thorough", false, "Check that all Runtastic activities are on Strava (not just all activities until the last recorded Strava activity)")

	errMissingCredentials = errors.New("Missing email address, password, or access token/client id and secret")
)

const usage = `Usage of Runtastic->Strava Importer:
  -email string
	Email (required)
		
  -password string
	Password (required)
		
  -token string
	Strava access token for your account
				
  -id string
	Strava API client ID
		
  -secret string
	Strava API client secret associated with the given ID. 
		
  -thorough
	Checks that all activities on the Runtastic account are also on Strava. Normal check goes only until the last recorded Strava activity
		
A token or client ID and secret must be provided. If a token is provided, it will be used. If a client ID and secret are used, you will be prompted to give this app permission to access your account.`

func getInput() (string, string, string, string, string, bool, error) {
	email := *email
	password := *password
	token := *token
	id := *id
	secret := *secret
	thorough := *thorough

	if token == "" {
		token = os.Getenv("STRAVA_TOKEN")
	}

	if token != "" || (id != "" && secret != "") {

		if email != "" && password != "" {
			return email, password, token, id, secret, thorough, nil
		}

		email = os.Getenv("RUNTASTIC_EMAIL")
		password = os.Getenv("RUNTASTIC_PASSWORD")

		if email != "" && password != "" {
			return email, password, token, id, secret, thorough, nil
		}

	}

	return "", "", "", "", "", false, errMissingCredentials
}

func main() {
	flag.Parse()

	email, password, token, id, secret, thorough, err := getInput()

	if err != nil {
		fmt.Println(usage)
		os.Exit(1)
	}

	if token == "" {
		token, err = oauth.GetToken(id, secret)
	}

	if err != nil {
		glog.Exit(err)
	}

	if thorough {
		fmt.Println("Thorough mode is on: All runtastic activities will be uploaded")
	}

	ctx := context.Background()
	session, err := api.Login(ctx, email, password)

	if err != nil {
		glog.Exit(err)
	}

	fmt.Println("Successfully logged into Runtastic")

	client := strava.NewClient(token)

	var count int

	switch thorough {
	case true:
		count, err = upload.UploadThorough(session, ctx, client)
	case false:
		count, err = upload.UploadNormal(session, ctx, client)
	}

	fmt.Printf("\nUploaded %d activities\n", count)

	if err != nil {
		glog.Exit(err)
	}

	glog.Flush()

}
