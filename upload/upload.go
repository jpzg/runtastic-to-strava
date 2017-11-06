package upload

import (
	"context"

	"github.com/metalnem/runtastic/api"
	"github.com/strava/go.strava"
)

func UploadNormal(session *api.Session, ctx context.Context, athlete *strava.CurrentAthleteService) (int, error) {
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
