// This oauth package allows the user to generate an access token with write permissions for the app to use. Based on https://github.com/strava/go.strava/blob/master/oauth.go
package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/strava/go.strava"
)

const port = 8080 // port of local demo server

var authenticator *strava.OAuthAuthenticator
var tokenChannel chan string // Icky global variables, I'm sure there's a better way but this works at the moment
var errorChannel chan error

func runServer(id string, secret string) {

	// define a strava.OAuthAuthenticator to hold state.
	// The callback url is used to generate an AuthorizationURL.
	// The RequestClientGenerator can be used to generate an http.RequestClient.
	// This is usually when running on the Google App Engine platform.
	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://localhost:%d/exchange_token", port),
		RequestClientGenerator: nil,
	}

	http.HandleFunc("/", indexHandler)

	path, err := authenticator.CallbackPath()
	if err != nil {
		// possibly that the callback url set above is invalid
		errorChannel <- err
		return
	}
	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	// start the server
	fmt.Printf("Visit http://localhost:%d/ to get an access token and continue the upload\n", port)
	fmt.Printf("ctrl-c to exit")
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<a href="%s">`, authenticator.AuthorizationURL("state1", strava.Permissions.Write, true))
	fmt.Fprint(w, `<div>Click here to connect with Strava!</div>`)
	fmt.Fprint(w, `</a>`)
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "SUCCESS:\nAccess token received\n")
	fmt.Fprintf(w, "State: %s\n\n", auth.State)
	fmt.Fprintf(w, "Access Token: %s\n\n", auth.AccessToken)

	fmt.Fprintf(w, "The Authenticated Athlete (you):\n")
	content, _ := json.MarshalIndent(auth.Athlete, "", " ")
	fmt.Fprint(w, string(content))

	close(errorChannel)
	tokenChannel <- auth.AccessToken
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "Write permissions were not authorized, uploading cannot continue\n")
		fmt.Println("Write permissions were not authorized, uploading cannot continue\n")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "Incorrect client_id or client_secret")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}

	errorChannel <- err
}

func GetToken(id string, secret string) (string, error) {
	tokenChannel = make(chan string)
	errorChannel = make(chan error)

	fmt.Println("No access token provided. Please follow the instructions to generate an access token.")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	}
	strava.ClientId = idInt
	strava.ClientSecret = secret

	go runServer(id, secret)

	err = <-errorChannel

	if err != nil {
		return "", err
	}

	token := <-tokenChannel

	return token, nil
}
