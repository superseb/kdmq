package util

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	sv "github.com/coreos/go-semver/semver"
	"github.com/hashicorp/go-retryablehttp"
	kd "github.com/rancher/rancher/pkg/controllers/management/kontainerdrivermetadata"
	ext "github.com/rancher/rancher/pkg/image/external"
	rketypes "github.com/rancher/rke/types"
	"github.com/rancher/rke/types/kdm"
)

func Difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func DifferenceOneWay(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func GetSemverFromString(version string) (semver.Version, error) {
	strippedVersion := strings.TrimPrefix(version, "v")
	semVersion, err := semver.Make(strippedVersion)
	if err != nil {
		return semver.Version{}, err
	}
	return semVersion, nil

}

func GetK8sVersionsForVersion(data kdm.Data, version string) ([]string, error) {
	linuxInfo, _ := kd.GetK8sVersionInfo(
		version,
		data.K8sVersionRKESystemImages,
		data.K8sVersionServiceOptions,
		data.K8sVersionWindowsServiceOptions,
		data.K8sVersionInfo,
	)
	var k8sVersions []string
	for k := range linuxInfo.RKESystemImages {
		k8sVersions = append(k8sVersions, k)
	}
	sort.Strings(k8sVersions)
	return k8sVersions, nil
}

func GetRKE2K8sVersionsForVersion(data kdm.Data, version string) ([]string, error) {
	var RKE2K8sVersions []string
	rke2AllImages, err := ext.GetExternalImages(version, data.RKE2, ext.RKE2, &sv.Version{
		Major: 1,
		Minor: 21,
		Patch: 0,
	})
	if err != nil {
		return RKE2K8sVersions, err
	}
	for _, image := range rke2AllImages {
		if strings.HasPrefix(image, "rancher/rke2-runtime") {
			splitImage := strings.Split(image, ":")
			if len(splitImage) == 2 {
				RKE2K8sVersions = append(RKE2K8sVersions, splitImage[1])
			}
		}
	}

	sort.Strings(RKE2K8sVersions)
	return RKE2K8sVersions, nil
}

func GetKDMDataFromURL(channel, channelVersion string) (kdm.Data, error) {
	metadataURL := fmt.Sprintf("https://releases.rancher.com/kontainer-driver-metadata/%s-%s/data.json", channel, channelVersion)
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	req, err := retryablehttp.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while calling NewRequest for [%s], error: %v", metadataURL, err)
	}
	resp, err := retryClient.Do(req)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while executing HTTP request, error: %v", err)
	}

	//resp, err := retryablehttp.Get(metadataURL)
	if err != nil || resp.StatusCode >= 400 {
		return kdm.Data{}, fmt.Errorf("error during HTTP get to [%s], error: %v", metadataURL, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error during reading HTTP response body, error: %v", err)
	}
	data, err := kdm.FromData(b)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error during translating response body to KDM data, error: %v", err)
	}
	return data, nil
}

func GetKDMDataFromEmbedded(rancherVersion string) (kdm.Data, error) {
	goModURL := fmt.Sprintf("https://raw.githubusercontent.com/rancher/rancher/%s/go.mod", rancherVersion)
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	req1, err := retryablehttp.NewRequest("GET", goModURL, nil)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while calling NewRequest for [%s], error: %v", goModURL, err)
	}
	resp1, err := retryClient.Do(req1)
	if err != nil || resp1.StatusCode >= 400 {
		return kdm.Data{}, fmt.Errorf("error during HTTP get to [%s], error: %v", goModURL, err)
	}

	defer resp1.Body.Close()
	goMod, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error during reading HTTP response body, error: %v", err)
	}
	re := regexp.MustCompile("(?m)github.com/rancher/rke\\s(.*)$")
	match := re.FindStringSubmatch(string(goMod))
	if len(match) < 1 {
		return kdm.Data{}, fmt.Errorf("error during finding RKE version from go.mod")
	}
	rkeVersion := match[1]
	splittedRKEVersion := strings.Split(rkeVersion, "-")
	if len(splittedRKEVersion) >= 3 {
		rkeVersion = splittedRKEVersion[2]
	}
	metadataURL := fmt.Sprintf("https://raw.githubusercontent.com/rancher/rke/%s/data/data.json", rkeVersion)
	req2, err := retryablehttp.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while calling NewRequest for [%s], error: %v", metadataURL, err)
	}
	resp2, err := retryClient.Do(req2)
	if err != nil || resp2.StatusCode >= 400 {
		return kdm.Data{}, fmt.Errorf("error during HTTP get to [%s], error: %v", goModURL, err)
	}

	defer resp2.Body.Close()
	b, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error during reading 2nd HTTP response body, error: %v", err)
	}
	data, err := kdm.FromData(b)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error during translating response body to KDM data, error: %v", err)
	}
	return data, nil
}

func GetUniqueSystemImageList(rkeSystemImages rketypes.RKESystemImages) []string {
	imagesReflect := reflect.ValueOf(rkeSystemImages)
	var images []string
	for i := 0; i < imagesReflect.NumField(); i++ {
		if imagesReflect.Field(i).Interface().(string) == "" {
			continue
		}
		images = append(images, imagesReflect.Field(i).Interface().(string))
	}
	return GetUniqueSlice(images)
}

func GetUniqueSlice(slice []string) []string {
	encountered := map[string]bool{}
	unqiue := []string{}

	for i := range slice {
		if encountered[slice[i]] {
			continue
		} else {
			encountered[slice[i]] = true
			unqiue = append(unqiue, slice[i])
		}
	}
	return unqiue
}

func IsValidChannel(channel string) (bool, error) {
	validChannels := []string{"dev", "release", "embedded"}
	for _, validChannel := range validChannels {
		if validChannel == channel {
			return true, nil
		}
	}
	return false, fmt.Errorf("not a valid channel [%s], valid options are [%s]", channel, strings.Join(validChannels, ","))
}

func IsValidChannelVersion(channelVersion string) (bool, error) {
	match, err := regexp.MatchString("^v2\\.[4-9]", channelVersion)
	if err != nil {
		return false, fmt.Errorf("not a valid channel version [%s], error: %v", channelVersion, err)
	}
	if !match {
		return false, fmt.Errorf("not a valid channel version [%s]", channelVersion)
	}
	return true, nil
}

func GetAddonNames(data map[string]map[string]string) []string {
	var k8sAddons []string
	for addon := range data {
		if addon == "templateKeys" {
			continue
		}
		k8sAddons = append(k8sAddons, addon)
	}
	return k8sAddons
}

func GetTemplate(data map[string]map[string]string, templateName, k8sVersion string) (string, string, error) {
	versionData := data[templateName]
	toMatch, err := semver.Make(k8sVersion[1:])
	if err != nil {
		return "", "", fmt.Errorf("k8sVersion not sem-ver %s %v", k8sVersion, err)
	}
	for k := range versionData {
		testRange, err := semver.ParseRange(k)
		if err != nil {
			return "", "", fmt.Errorf("range for %s not sem-ver %v %v", templateName, testRange, err)
		}
		if testRange(toMatch) {
			return versionData[k], data[kdm.TemplateKeys][versionData[k]], nil
		}
	}
	return "", "", fmt.Errorf("no %s template found for k8sVersion %s", templateName, k8sVersion)
}
