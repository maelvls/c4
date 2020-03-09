package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

// This struct is only used for returning which VMs have been deleted so
// that we can send a nice Slack message.
type vm struct {
	name     string // e.g. content of tag:Name on aws, name in gcp & openstack
	provider string // e.g.: gcp, aws, openstack
	region   string // e.g.: eu-west1 (aws), europe-west-1a (gcp), UK1 (openstack)
	id       string // Is empty on gcp.
}

// The tenant name is the same as project name. The domain name is very
// often just "Default" in most OpenStack instances. The region is e.g.
// "UK1" depending on your OpenStack instance.
func nukeOpenStackInstances(region, authURL, domainName, username, password, projectName string, regex *regexp.Regexp, dryRun bool, olderThan time.Duration) ([]servers.Server, error) {
	providerClient, err := openstack.NewClient(authURL)

	if err != nil {
		return nil, fmt.Errorf("could not create an openstack client: %w", err)
	}

	err = openstack.Authenticate(providerClient, gophercloud.AuthOptions{
		IdentityEndpoint: authURL,
		DomainName:       domainName,
		Username:         username,
		Password:         password,
		TenantName:       projectName,
		AllowReauth:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: gophercloud.AvailabilityPublic,
	})
	if err != nil {
		return nil, err
	}

	p, err := servers.List(computeClient, servers.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}
	servs, err := servers.ExtractServers(p)
	if err != nil {
		return nil, err
	}

	var toBeDeleted []servers.Server
	for _, s := range servs {
		if !regex.MatchString(s.Name) {
			continue
		}

		age := osAge(s)
		if age >= olderThan {
			toBeDeleted = append(toBeDeleted, s)
			fmt.Printf("found openstack instance %s (%s), removing since age is %s\n", yel(s.Name), s.ID, red(age.String()))
		} else {
			fmt.Printf("found openstack instance %s (%s), keeping it since age is %s\n", yel(s.Name), s.ID, green(age.String()))
		}
	}

	if dryRun {
		return toBeDeleted, nil
	}

	for _, s := range toBeDeleted {
		err := servers.Delete(computeClient, s.ID).ExtractErr()
		if err != nil {
			return nil, err
		}
	}

	return toBeDeleted, nil
}

func osAge(s servers.Server) time.Duration {
	return time.Now().Sub(s.Created)
}
