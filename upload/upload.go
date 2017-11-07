package upload

import (
	"bytes"
	"context"
	"fmt"

	"github.com/metalnem/runtastic/api"
	"github.com/metalnem/runtastic/gpx"
	"github.com/strava/go.strava"
)

// initialize performs code that happens regardless of thoroughness, including getting metadata from runtastic and creating the Strava upload service.
func initialize(session *api.Session, ctx context.Context, client *strava.Client) ([]api.Metadata, *strava.UploadsService, *strava.CurrentAthleteService, error) {

	runtasticMetadata, err := session.GetMetadata(ctx)

	if err != nil {
		return nil, nil, nil, err
	}

	uploadsService := strava.NewUploadsService(client)
	athlete := strava.NewCurrentAthleteService(client)

	return runtasticMetadata, uploadsService, athlete, err
}

// uploadActivity takes a runtastic activity ID and uploads the corresponding activity to Strava.
func uploadActivity(id api.ActivityID, session *api.Session, ctx context.Context, uploadsService *strava.UploadsService) (*strava.UploadSummary, error) {

	var buffer bytes.Buffer
	exp := gpx.NewExporter(&buffer)

	activity, err := session.GetActivity(ctx, id)

	if err != nil {
		return nil, err
	}

	err = exp.Export(*activity)

	if err != nil {
		return nil, err
	}

	var activityType strava.ActivityType

	switch activity.Type {
	case "Running":
		activityType = strava.ActivityTypes.Run
	case "Biking":
		activityType = strava.ActivityTypes.Ride
	case "Swimming":
		activityType = strava.ActivityTypes.Swim
	case "Walking":
		activityType = strava.ActivityTypes.Walk
	default:
		activityType = strava.ActivityTypes.Workout
	}

	summary, err := uploadsService.Create("gpx", "runtastic_activity.gpx", &buffer).
		ActivityType(activityType).
		Description(activity.Notes).
		Do()

	return summary, err
}

func UploadNormal(session *api.Session, ctx context.Context, client *strava.Client) (int, error) {
	count := 0

	runtasticMetadata, uploadsService, athlete, err := initialize(session, ctx, client)

	if err != nil {
		return count, err
	}

	stravaMeta, err := athlete.ListActivities().Page(1).Do()

	if err != nil {
		return count, err
	}

	lastActivity := stravaMeta[0].StartDate
	fmt.Printf("Last strava activity was on %s\n", lastActivity)

	for i, _ := range runtasticMetadata {
		index := len(runtasticMetadata) - 1 - i

		if runtasticMetadata[index].StartTime.After(lastActivity) {
			summary, err := uploadActivity(runtasticMetadata[index].ID, session, ctx, uploadsService)

			if err != nil {
				return count, err
			}
			fmt.Println(summary)
			count++
		} else {
			break
		}
	}

	return count, nil
}

func UploadThorough(session *api.Session, ctx context.Context, client *strava.Client) (int, error) {
	count := 0

	//runtasticMetadata, uploadsService, athlete, err := initialize(session, ctx, client)
	_, _, _, err := initialize(session, ctx, client)

	if err != nil {
		return count, err
	}

	/*stravaMeta, err := athlete.ListActivities().Page(1).Do()

	if err != nil {
		return count, nil
	}
	*/
	return count, nil
}
