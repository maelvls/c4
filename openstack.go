package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

// The tenant name is the same as project name. The domain name is very
// often just "Default" in most OpenStack instances. The region is e.g.
// "UK1" depending on your OpenStack instance.
func nukeOpenStackInstances(region, authURL, domainName, username, password, projectName, nameContains string, dryRun bool) error {
	providerClient, err := openstack.NewClient(authURL)

	if err != nil {
		return fmt.Errorf("could not create an openstack client: %w", err)
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
		return fmt.Errorf("authentication failed: %w", err)
	}

	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: gophercloud.AvailabilityPublic,
	})
	if err != nil {
		return err
	}

	p, err := servers.List(computeClient, servers.ListOpts{}).AllPages()
	if err != nil {
		return err
	}
	servs, err := servers.ExtractServers(p)
	if err != nil {
		return err
	}

	var ids []string
	for _, s := range servs {
		if !strings.Contains(s.Name, nameContains) {
			continue
		}

		age := time.Now().Sub(s.Created)
		if age >= *olderThan {
			ids = append(ids, s.ID)
			fmt.Printf("found openstack instance %s (%s), removing since age is %s\n", yel(s.Name), s.ID, red(age.String()))
		} else {
			fmt.Printf("found openstack instance %s (%s), keeping it since age is %s\n", yel(s.Name), s.ID, green(age.String()))
		}
	}

	if dryRun {
		return nil
	}
	for _, id := range ids {
		err := servers.Delete(computeClient, id).ExtractErr()
		if err != nil {
			return err
		}
	}

	return nil
}
