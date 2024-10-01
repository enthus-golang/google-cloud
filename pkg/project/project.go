package project

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/compute/metadata"
)

// ProjectID returns the Google project ID based on the environment or the metadata server.
func ProjectID(ctx context.Context) (string, error) {
	if googleCloudProject := os.Getenv("GOOGLE_CLOUD_PROJECT"); googleCloudProject != "" {
		// Cloud Shell and App Engine set this environment variable to the project
		// ID, so use it if present. In case of App Engine the project ID is also
		// available from the GCE metadata server, but by using the environment
		// variable saves one request to the metadata server. The environment
		// project ID is only used if no project ID is provided in the
		// configuration.
		return googleCloudProject, nil
	}

	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("project id: %w", err)
	}

	return projectID, nil
}
