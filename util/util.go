package util

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	sv "github.com/coreos/go-semver/semver"
	"github.com/hashicorp/go-retryablehttp"
	mVersion "github.com/mcuadros/go-version"
	kd "github.com/rancher/rancher/pkg/controllers/management/kontainerdrivermetadata"
	ext "github.com/rancher/rancher/pkg/image/external"
	rketypes "github.com/rancher/rke/types"
	"github.com/rancher/rke/types/kdm"
	rkeutil "github.com/rancher/rke/util"
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

func GetK3SK8sVersionsForVersion(data kdm.Data, version string) ([]string, error) {
	var K3SK8sVersions []string
	k3sAllImages, err := ext.GetExternalImages(version, data.K3S, ext.K3S, &sv.Version{
		Major: 1,
		Minor: 21,
		Patch: 0,
	})

	if err != nil {
		return K3SK8sVersions, err
	}
	for _, image := range k3sAllImages {
		if strings.HasPrefix(image, "rancher/k3s-upgrade") {
			splitImage := strings.Split(image, ":")
			if len(splitImage) == 2 {
				K3SK8sVersions = append(K3SK8sVersions, splitImage[1])
			}
		}
	}

	sort.Strings(K3SK8sVersions)
	return K3SK8sVersions, nil
}

func GetKDMDataFromFile(file string) (kdm.Data, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("Error while trying to read file [%s], error: %v", file, err)
	}
	data, err := kdm.FromData(b)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error translating file data to KDM data, error: %v", err)
	}
	return data, nil
}

func GetKDMDataFromCustomURL(url string) (kdm.Data, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while calling NewRequest for [%s], error: %v", url, err)
	}
	resp, err := retryClient.Do(req)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while executing HTTP request, error: %v", err)
	}

	if err != nil || resp.StatusCode >= 400 {
		return kdm.Data{}, fmt.Errorf("error during HTTP get to [%s], error: %v", url, err)
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

func GetKDMDataFromURL(channel, channelVersion string) (kdm.Data, error) {
	if channel == "latest" {
		channel = "release"
	}
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
	metadataURL := fmt.Sprintf("https://raw.githubusercontent.com/superseb/kdmq/main/embedded/data.%s.json", rancherVersion)
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	req, err := retryablehttp.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return kdm.Data{}, fmt.Errorf("error while calling NewRequest for [%s], error: %v", metadataURL, err)
	}
	resp, err := retryClient.Do(req)
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
	validChannels := []string{"release", "latest", "dev"}
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
	sort.Strings(k8sAddons)
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

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetDataForChannel(version, channel string) (kdm.Data, error) {
	var data kdm.Data
	var err error
	if strings.HasPrefix(channel, "./") {
		fileExists, err := FileExists(channel)
		if err != nil {
			return data, fmt.Errorf("Error checking local data file: [%s], error [%v]", channel, err)
		}
		if !fileExists {
			return data, fmt.Errorf("Local data file [%s] does not exist", channel)
		}
		data, err = GetKDMDataFromFile(channel)
		if err != nil {
			return data, fmt.Errorf("Error while trying to get KDM data from local data file, error [%v]", err)
		}
		return data, nil
	} else if isValidUrl(channel) {
		data, err = GetKDMDataFromCustomURL(channel)
		if err != nil {
			return data, fmt.Errorf("Error while trying to get KDM data from custom URL, error [%v]", err)
		}
		return data, nil
	} else {
		validChannel, err := IsValidChannel(channel)
		if !validChannel {
			return data, fmt.Errorf("Not a valid channel: [%s], error [%v]", channel, err)
		}

		if channel == "release" {
			data, err = GetKDMDataFromEmbedded(version)
			if err != nil {
				return data, fmt.Errorf("Error while trying to get KDM data from embedded, error [%v]", err)
			}
			return data, nil
		} else {

			semVersion, err := GetSemverFromString(version)
			if err != nil {
				return data, fmt.Errorf("Not a valid semver version: [%s], error [%v]", version, err)
			}

			data, err = GetKDMDataFromURL(channel, fmt.Sprintf("v%d.%d", semVersion.Major, semVersion.Minor))
			if err != nil {
				return data, fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
			}
			return data, nil
		}
	}
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func GetLatestMajorMinorK8sVersions(k8sVersions []string) []string {
	maxVersionForMajorK8sVersion := map[string]string{}
	latestK8sVersions := []string{}
	for _, k8sVersion := range k8sVersions {
		majorVersion := rkeutil.GetTagMajorVersion(k8sVersion)
		if curr, ok := maxVersionForMajorK8sVersion[majorVersion]; !ok || mVersion.Compare(k8sVersion, curr, ">") {
			maxVersionForMajorK8sVersion[majorVersion] = k8sVersion
		}
	}
	for _, latestK8sVersion := range maxVersionForMajorK8sVersion {
		latestK8sVersions = append(latestK8sVersions, latestK8sVersion)
	}

	return latestK8sVersions
}

func GetValidProducts() []string {
	return []string{"rke", "rke2", "k3s"}
}

func PrependV(version string) string {
	if !strings.HasPrefix(version, "v") {
		return fmt.Sprintf("v%s", version)
	}
	return version
}
