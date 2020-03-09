package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func nukeGCPInstances(jsonKey, nameContains string, dryRun bool, olderThan time.Duration) error {
	type cred struct {
		ProjectID string `json:"project_id"`
	}
	var c cred
	err := json.Unmarshal([]byte(jsonKey), &c)
	if err != nil {
		return fmt.Errorf("while reading json key: %w", err)
	}

	project := c.ProjectID

	jwt, err := google.JWTConfigFromJSON([]byte(jsonKey), compute.CloudPlatformScope)
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, jwt.TokenSource(ctx))
	computeClient, err := compute.NewService(ctx, option.WithHTTPClient(httpClient))

	resp, err := computeClient.Instances.AggregatedList(project).Do()
	if err != nil {
		return fmt.Errorf("while listing instances: %w", err)
	}

	var toBeDeleted []*compute.Instance
	for _, instances := range resp.Items {
		for _, inst := range instances.Instances {
			if !strings.Contains(inst.Name, nameContains) {
				continue
			}

			creation, err := time.Parse(time.RFC3339, inst.CreationTimestamp)
			if err != nil {
				return fmt.Errorf("parsing timestamp for instance %s: %w", inst.Name, err)
			}

			age := time.Now().Sub(creation)
			if age >= olderThan {
				toBeDeleted = append(toBeDeleted, inst)
				fmt.Printf("found gcp instance %s (%s), removing since age is %s\n", yel(inst.Name), shorterGCPURL(inst.Zone), red(age.String()))
			} else {
				fmt.Printf("found gcp instance %s (%s), keeping it since age is %s\n", yel(inst.Name), shorterGCPURL(inst.Zone), green(age.String()))
			}
		}
	}

	if dryRun {
		return nil
	}
	for _, instance := range toBeDeleted {
		_, err := computeClient.Instances.Delete(project, shorterGCPURL(instance.Zone), instance.Name).Do()
		if err != nil {
			return fmt.Errorf("while deleting %s: %w", instance.Name, err)
		}
	}

	return nil
}

// The GCP API uses long URLs as identifiers. For example, an instance zone
// will be given as:
//
//  https://www.googleapis.com/compute/v1/projects/project-1/zones/europe-west2-a.
//
// This function returns the last segment of the URL. In this example, that
// would return "europe-west2-a".
func shorterGCPURL(URL string) string {
	elmts := strings.Split(URL, "/")
	return elmts[len(elmts)-1]
}
