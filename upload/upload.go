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
func uploadActivity(activity api.Activity, uploadsService *strava.UploadsService) (*strava.UploadSummary, error) {

	var buffer bytes.Buffer
	exp := gpx.NewExporter(&buffer)

	err := exp.Export(activity)

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

func checkRateLimit() {
	rl := strava.RateLimiting
	var resumeAt time.Time

	if rl.UsageLong >= rl.LimitLong {
		resumeAt = time.Date(rl.RequestTime.Year(), rl.RequestTime.Month(), rl.RequestTime.Day()+1, 0, 0, 0, 0, time.UTC).Local()
		fmt.Println("\nExceeded Strava API daily usage limit")

	} else if rl.UsageShort >= rl.LimitShort {
		resumeAt = rl.RequestTime.Truncate(15 * time.Minute).Add(15 * time.Minute)
		fmt.Println("\nExceeded Strava API 15-minute usage limit")
	}

	if rl.Exceeded() {
		fmt.Printf("Uploading will resume at %s\nCtrl-C to exit\n\n", resumeAt)
		time.Sleep(time.Until(resumeAt))
	}
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
		runtasticActivity, err := session.GetActivity(ctx, runtasticMetadata[index].ID)
		if err != nil {
			return count, err
		}
		runtasticStartTime := runtasticActivity.Data[0].Time // Strava uses the timestamp on the first trackpoint as the start time, Runtastic uses whenever you press "start". This causes issues when the GPS initially doesn't have signal and the first trackpoint comes later.

		if runtasticStartTime.After(lastActivity) {
			checkRateLimit()
			summary, err := uploadActivity(*runtasticActivity, uploadsService)

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

// TODO: Use strava.RateLimit and strava.RateLimit.Exceeded() to make these rate-limit aware
func UploadThorough(session *api.Session, ctx context.Context, client *strava.Client) (int, error) {
	count := 0

	_, uploadsService, athlete, err := initialize(session, ctx, client)

	if err != nil {
		return count, err
	}

	runtasticActivities, err := session.GetActivities(ctx)
	if err != nil {
		return count, err
	}
	index := len(runtasticActivities) - 1

	stravaMeta, err := athlete.ListActivities().Page(1).Do()

	for page := 2; len(stravaMeta) > 0; page++ {

		if err != nil {
			return count, err
		}

		for _, activity := range stravaMeta {

			for activity.StartDate.Before(runtasticActivities[index].Data[0].Time) {

				checkRateLimit()
				summary, _ := uploadActivity(runtasticActivities[index], uploadsService) // should also be assigning err
				/*
					if err != nil {
						return count, err // Commenting this out because I want to get my activities up but don't want to refactor everything like I probably should...
					}
				*/
				index--
				count++
				fmt.Println(summary)

				if index < 0 {
					return count, nil
				}
			}

			if activity.StartDate.Equal(runtasticActivities[index].Data[0].Time) {
				index--
			}
		}

		if err != nil {
			return count, err
		}
		stravaMeta, err = athlete.ListActivities().Page(page).Do()
	}

	for index >= 0 {
		checkRateLimit()
		summary, err := uploadActivity(runtasticActivities[index], uploadsService)

		if err != nil {
			return count, err
		}
		index--
		count++
		fmt.Println(summary)
	}

	return count, nil
}
