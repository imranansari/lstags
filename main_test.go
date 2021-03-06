/*

NB!
NB! "main" package tests are only for integration testing
NB! "main" package is bare and all unit tests are put into packages
NB!

*/

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ivanilves/lstags/api/v1"
	"github.com/ivanilves/lstags/api/v1/registry"
)

func runEnd2EndJob(pullRefs, seedRefs []string) ([]string, error) {
	apiConfig := v1.Config{}

	api, err := v1.New(apiConfig)
	if err != nil {
		return nil, err
	}

	collection, err := api.CollectTags(pullRefs)
	if err != nil {
		return nil, err
	}

	registryContainer, err := registry.LaunchContainer()
	if err != nil {
		return nil, err
	}

	defer registryContainer.Destroy()

	if len(seedRefs) > 0 {
		if _, err := registryContainer.SeedWithImages(seedRefs...); err != nil {
			return nil, err
		}
	}

	pushConfig := v1.PushConfig{Registry: registryContainer.Hostname()}

	pushCollection, err := api.CollectPushTags(collection, pushConfig)
	if err != nil {
		return nil, err
	}

	return pushCollection.TaggedRefs(), nil
}

func TestEnd2End(t *testing.T) {
	var testCases = []struct {
		pullRefs         []string
		seedRefs         []string
		expectedPushRefs []string
		isCorrect        bool
	}{
		{
			[]string{},
			[]string{"alpine:3.7", "busybox:latest"},
			[]string{},
			false,
		},
		{
			[]string{"alpine:3.7", "busybox:latest"},
			[]string{"alpine:3.7", "busybox:latest"},
			[]string{},
			true,
		},
		{
			[]string{"alpine:3.7", "busybox:latest"},
			[]string{"alpine:3.7", "quay.io/calico/ctl:v1.6.1"},
			[]string{"busybox:latest"},
			true,
		},
		{
			[]string{"alpine:3.7", "busybox:latest", "gcr.io/google_containers/pause-amd64:3.0"},
			[]string{"alpine:3.7"},
			[]string{"busybox:latest", "gcr.io/google_containers/pause-amd64:3.0"},
			true,
		},
		{
			[]string{"idonotexist:latest", "busybox:latest"},
			[]string{},
			[]string{},
			false,
		},
		{
			[]string{"busybox:latest"},
			[]string{"idonotexist:latest"},
			[]string{},
			false,
		},
		{
			[]string{"busybox:latest", "!@#$%^&*"},
			[]string{},
			[]string{},
			false,
		},
		{
			[]string{"alpine:3.7", "busybox:latest"},
			[]string{"!@#$%^&*", "alpine:3.7"},
			[]string{},
			false,
		},
	}

	assert := assert.New(t)

	for _, testCase := range testCases {
		pushRefs, err := runEnd2EndJob(testCase.pullRefs, testCase.seedRefs)

		if testCase.isCorrect {
			assert.Nil(err, "should be no error")
		} else {
			assert.NotNil(err, "should be an error")
		}

		if err != nil {
			continue
		}

		assert.Equal(testCase.expectedPushRefs, pushRefs, fmt.Sprintf("%+v", testCase))
	}
}
