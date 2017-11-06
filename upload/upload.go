package upload

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/metalnem/runtastic/api"
	"github.com/metalnem/runtastic/gpx"
	"github.com/strava/go.strava"
)

// initialize performs code that happens regardless of thoroughness, including getting metadata from runtastic and creating the Strava upload service.
func initialize(session *api.Session, ctx context.Context, athlete *strava.CurrentAthleteService) ([]api.Metadata, *strava.UploadsService, error) {

	runtastic_metadata, err := session.GetMetadata(ctx)

	if err != nil {
		return nil, nil, nil, err
	}

}

// uploadActivity takes a runtastic activity ID and uploads the corresponding activity to Strava.
func uploadActivity(id api.ActivityID, session *api.Session, ctx context.Context, uploadsService *strava.UploadsService) (*strava.UploadSummary, error) {

	var buffer bytes.Buffer
	exp := gpx.NewExporter(buffer)

	activity, err := session.GetActivity(ctx, id)

	if err != nil {
		return nil, err
	}

	err = exp.Export(activity)

	if err != nil {
		return nil, err
	}

	activityType := activity.Type

	switch activityType {
	case "Running":
		activityType = "run"
	case "Biking":
		activityType = "ride"
	case "Swimming":
		activityType = "swim"
	case "Walking":
		activityType = "walk"
	default:
		activityType = "workout"
	}

	summary, err := uploadsService.Create("gpx", "runtastic_activity.gpx", buffer).
		ActivityType(activityType).
		Description(activity.Notes).
		Do()

	return summary, err
}

func UploadNormal(session *api.Session, ctx context.Context, athlete *strava.CurrentAthleteService) (int, error) {
	count := 0
	strava_meta, err := athlete.ListActivities().Page(1).Do()

	if err != nil {
		return count, err
	}
	last_activity := strava_meta[0].StartDate
	fmt.Printf("Last strava activity was on %s\n", last_activity)

	runtastic_meta, err := session.GetMetadata(ctx)

	if err != nil {
		return count, err
	}
	fmt.Println("Successfully logged into Runtastic")

	return count, nil
}

func UploadThorough(session *api.Session, ctx context.Context, athlete *strava.CurrentAthleteService) (int, error) {
	count := 0
	strava_meta, err := athlete.ListActivities().Page(1).Do()

	if err != nil {
		return count, nil
	}

	runtastic_meta, err := session.GetMetadata(ctx)

	if err != nil {
		return count, nil
	}

	return count, nil
}
