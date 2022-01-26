package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/rancher/rke/types/kdm"
	"github.com/superseb/kdmq/util"
	"github.com/urfave/cli/v2"
)

var (
	Version = "v0.0.0-dev"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "listk8s",
				Aliases: []string{"lk"},
				Usage:   "list k8s versions for Rancher version",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <rancher_version> <channel>", c.Command.FullName())

					if c.Args().Len() < 2 {
						return fmt.Errorf("Not enough parameters\n%s", commandUsage)
					}

					version := c.Args().Get(0)
					channel := c.Args().Get(1)

					var data kdm.Data
					var err error

					data, err = util.GetDataForChannel(version, channel)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}
					k8sVersions, err := util.GetK8sVersionsForVersion(data, version)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions, error [%v]", err)
					}

					if c.Bool("verbose") {
						fmt.Printf("Kubernetes versions found for version [%s] in channel [%s]:\n%s\n", version, channel, strings.Join(k8sVersions, "\n"))
						return nil
					}
					fmt.Printf("%s\n", strings.Join(k8sVersions, "\n"))
					return nil

				},
			},
			{
				Name:    "diffk8s",
				Aliases: []string{"dk"},
				Usage:   "diff 2 k8s version",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <rancher_version1> <rancher_version2> <channel1> [channel2]", c.Command.FullName())

					if c.Args().Len() < 3 {
						return fmt.Errorf("Not enough parameters\n:%s", commandUsage)
					}
					version1 := c.Args().Get(0)
					version2 := c.Args().Get(1)
					channel1 := c.Args().Get(2)
					var channel2 string
					if c.Args().Len() == 4 {
						channel2 = c.Args().Get(3)
					}

					validChannel, err := util.IsValidChannel(channel1)
					if !validChannel {
						return fmt.Errorf("Not a valid channel1: [%s], error [%v]", channel1, err)
					}
					var dataVersion1 kdm.Data
					var dataVersion2 kdm.Data
					var customChannel2 bool

					dataVersion1, err = util.GetDataForChannel(version1, channel1)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}
					if channel2 != "" {
						customChannel2 = true
						dataVersion2, err = util.GetDataForChannel(version2, channel2)
						if err != nil {
							return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
						}

					}
					var k8sVersionsVersion2 []string

					k8sVersionsVersion1, err := util.GetK8sVersionsForVersion(dataVersion1, version1)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions for [%s], error: [%v]", version1, err)
					}

					if customChannel2 {
						k8sVersionsVersion2, err = util.GetK8sVersionsForVersion(dataVersion2, version2)
						if err != nil {
							return fmt.Errorf("Error while trying to get k8s versions for [%s], error: [%v]", version2, err)
						}
					} else {
						k8sVersionsVersion2, err = util.GetK8sVersionsForVersion(dataVersion1, version2)
						if err != nil {
							return fmt.Errorf("Error while trying to get k8s versions for [%s], error: [%v]", version2, err)
						}

					}
					var diffK8sVersions []string
					if c.Bool("diff-oneway") {
						diffK8sVersions = util.DifferenceOneWay(k8sVersionsVersion2, k8sVersionsVersion1)
					} else {
						diffK8sVersions = util.Difference(k8sVersionsVersion1, k8sVersionsVersion2)
					}
					sort.Strings(diffK8sVersions)

					replyMessage := fmt.Sprintf("Kubernetes versions found for version [%s] in channel [%s]:\n\n%s\n", version1, channel1, strings.Join(k8sVersionsVersion1, "\n"))
					if customChannel2 {
						replyMessage = fmt.Sprintf("%s\nKubernetes versions found for version [%s] in channel [%s]:\n\n%s\n", replyMessage, version2, channel2, strings.Join(k8sVersionsVersion2, "\n"))
					} else {
						replyMessage = fmt.Sprintf("%s\nKubernetes versions found for version [%s] in channel [%s]:\n\n%s\n", replyMessage, version2, channel1, strings.Join(k8sVersionsVersion2, "\n"))
					}
					if c.Bool("verbose") {
						replyMessage = fmt.Sprintf("%s\nDifference:\n%s\n\n", replyMessage, strings.Join(diffK8sVersions, "\n"))
						fmt.Printf(replyMessage)
						return nil
					}
					fmt.Printf("%s\n", strings.Join(diffK8sVersions, "\n"))

					return nil

				},
			},
			{
				Name:    "diffallk8simages",
				Aliases: []string{"daki"},
				Usage:   "diff all images between two Rancher versions",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <rancher_version1> <rancher_version2> <channel1> <channel2>", c.Command.FullName())

					if c.Args().Len() < 4 {
						return fmt.Errorf("Not enough parameters\n:%s", commandUsage)
					}
					version1 := c.Args().Get(0)
					version2 := c.Args().Get(1)
					channel1 := c.Args().Get(2)
					channel2 := c.Args().Get(3)

					var dataVersion1 kdm.Data
					var dataVersion2 kdm.Data
					var err error

					dataVersion1, err = util.GetDataForChannel(version1, channel1)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}

					dataVersion2, err = util.GetDataForChannel(version2, channel2)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}

					k8sVersionsVersion1, err := util.GetK8sVersionsForVersion(dataVersion1, version1)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions for [%s], error: [%v]", version1, err)
					}

					k8sVersionsVersion2, err := util.GetK8sVersionsForVersion(dataVersion2, version2)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions for [%s], error: [%v]", version1, err)
					}

					var k8sImagesVersion1 []string
					var k8sImagesVersion2 []string

					for _, k8sImageVersion1 := range k8sVersionsVersion1 {
						k8sImagesVersion1 = append(k8sImagesVersion1, util.GetUniqueSystemImageList(dataVersion1.K8sVersionRKESystemImages[k8sImageVersion1])...)
					}

					for _, k8sImageVersion2 := range k8sVersionsVersion2 {
						k8sImagesVersion2 = append(k8sImagesVersion2, util.GetUniqueSystemImageList(dataVersion2.K8sVersionRKESystemImages[k8sImageVersion2])...)
					}

					var diffK8sImageVersions []string
					if c.Bool("diff-oneway") {
						diffK8sImageVersions = util.DifferenceOneWay(k8sImagesVersion2, k8sImagesVersion1)
					} else {
						diffK8sImageVersions = util.Difference(k8sImagesVersion1, k8sImagesVersion2)
					}
					diffK8sImageVersions = util.GetUniqueSlice(diffK8sImageVersions)
					sort.Strings(diffK8sImageVersions)

					replyMessage := fmt.Sprintf("Images found for version [%s] in channel [%s]:\n\n%s\n", version1, channel1, strings.Join(k8sImagesVersion1, "\n"))
					replyMessage = fmt.Sprintf("%s\nImages found for version [%s] in channel [%s]:\n\n%s\n", replyMessage, version2, channel2, strings.Join(k8sImagesVersion2, "\n"))
					if c.Bool("verbose") {
						replyMessage = fmt.Sprintf("%s\nDifference:\n%s\n\n", replyMessage, strings.Join(diffK8sImageVersions, "\n"))
						fmt.Printf(replyMessage)
						return nil
					}
					fmt.Printf("%s\n", strings.Join(diffK8sImageVersions, "\n"))

					return nil

				},
			},
			{
				Name:    "listk8simages",
				Aliases: []string{"lki"},
				Usage:   "list k8s images",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <k8s_version> <rancher_version> <channel>", c.Command.FullName())
					if c.Args().Len() < 3 {
						return fmt.Errorf("Not enough parameters\n%s", commandUsage)
					}
					k8sVersion := c.Args().Get(0)
					version := c.Args().Get(1)
					channel := c.Args().Get(2)

					var data kdm.Data
					var err error

					data, err = util.GetDataForChannel(version, channel)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}

					k8sVersions, err := util.GetK8sVersionsForVersion(data, version)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions, error [%v]", err)
					}

					uniqueImages := util.GetUniqueSystemImageList(data.K8sVersionRKESystemImages[k8sVersion])
					if len(uniqueImages) == 0 {
						return fmt.Errorf("No images found for [%s] and Rancher version [%s] and channel [%s], latest available k8s patch versions: [%s]", k8sVersion, version, channel, strings.Join(k8sVersions, ","))
					}

					if c.Bool("verbose") {
						fmt.Printf("Images for Kubernetes version [%s] for channel [%s]:\n\n%s\n", k8sVersion, channel, strings.Join(uniqueImages, "\n"))
						return nil
					}
					fmt.Printf("%s\n", strings.Join(uniqueImages, "\n"))
					return nil

				},
			},
			{
				Name:    "diffk8simages",
				Aliases: []string{"dki"},
				Usage:   "diff 2 k8s version images",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <k8s_version1> <k8s_version2> <rancher_version> <channel1> <channel2>", c.Command.FullName())
					if c.Args().Len() < 5 {
						return fmt.Errorf("Not enough parameters\n%s", commandUsage)
					}
					k8sVersion1 := c.Args().Get(0)
					k8sVersion2 := c.Args().Get(1)
					version := c.Args().Get(2)
					channel1 := c.Args().Get(3)
					channel2 := c.Args().Get(4)

					var dataVersion1 kdm.Data
					var dataVersion2 kdm.Data
					var err error

					dataVersion1, err = util.GetDataForChannel(version, channel1)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}
					dataVersion2, err = util.GetDataForChannel(version, channel2)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}

					uniqueImagesK8sVersion1 := util.GetUniqueSystemImageList(dataVersion1.K8sVersionRKESystemImages[k8sVersion1])
					lenUniqueImagesK8sVersion1 := len(uniqueImagesK8sVersion1)
					uniqueImagesK8sVersion2 := util.GetUniqueSystemImageList(dataVersion2.K8sVersionRKESystemImages[k8sVersion2])
					lenUniqueImagesK8sVersion2 := len(uniqueImagesK8sVersion2)

					k8sVersions1, err := util.GetK8sVersionsForVersion(dataVersion1, version)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions, error [%v]", err)
					}

					k8sVersions2, err := util.GetK8sVersionsForVersion(dataVersion2, version)
					if err != nil {
						return fmt.Errorf("Error while trying to get k8s versions, error [%v]", err)
					}

					if lenUniqueImagesK8sVersion1 == 0 {
						return fmt.Errorf("No images found for [%s] and Rancher version [%s] and channel [%s], latest available k8s patch versions: [%s]", k8sVersion1, version, channel1, strings.Join(k8sVersions1, ","))
					}
					if lenUniqueImagesK8sVersion2 == 0 {
						return fmt.Errorf("No images found for [%s] and Rancher version [%s] and channel [%s], latest available k8s patch versions: [%s]", k8sVersion2, version, channel2, strings.Join(k8sVersions2, ","))
					}

					var diffImages []string
					if c.Bool("diff-oneway") {
						diffImages = util.DifferenceOneWay(uniqueImagesK8sVersion2, uniqueImagesK8sVersion1)
					} else {
						diffImages = util.Difference(uniqueImagesK8sVersion1, uniqueImagesK8sVersion2)
					}

					replyMessage := fmt.Sprintf("Images [%d] for Kubernetes version [%s] for channel [%s]:\n\n%s\n", lenUniqueImagesK8sVersion1, k8sVersion1, channel1, strings.Join(uniqueImagesK8sVersion1, "\n"))
					replyMessage = fmt.Sprintf("%s\nImages [%d] for Kubernetes version [%s] for channel [%s]:\n\n%s\n", replyMessage, lenUniqueImagesK8sVersion2, k8sVersion2, channel2, strings.Join(uniqueImagesK8sVersion2, "\n"))
					replyMessage = fmt.Sprintf("%s\nDifference:\n%s\n", replyMessage, strings.Join(diffImages, "\n"))

					if c.Bool("verbose") {
						fmt.Printf(replyMessage)
						return nil
					}
					fmt.Printf("%s\n", strings.Join(diffImages, "\n"))
					return nil

				},
			},
			{
				Name:    "listk8saddons",
				Aliases: []string{"lka"},
				Usage:   "list k8s version addons",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <k8s_version> <rancher_version> <channel>", c.Command.FullName())
					if c.Args().Len() < 3 {
						return fmt.Errorf("Not enough parameters\n%s", commandUsage)
					}
					k8sVersion := c.Args().Get(0)
					version := c.Args().Get(1)
					channel := c.Args().Get(2)

					var data kdm.Data
					var err error

					data, err = util.GetDataForChannel(version, channel)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}

					k8sAddons := util.GetAddonNames(data.K8sVersionedTemplates)

					replyMessage := fmt.Sprintf("Addons for Kubernetes version [%s] for channel [%s]:\n", k8sVersion, channel)

					tableString := &strings.Builder{}
					table := tablewriter.NewWriter(tableString)

					if c.Bool("verbose") {
						table.SetHeader([]string{"Addon", "Template name"})
					}
					table.SetAutoWrapText(false)
					table.SetAutoFormatHeaders(true)
					table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
					table.SetAlignment(tablewriter.ALIGN_LEFT)
					table.SetCenterSeparator("")
					table.SetColumnSeparator("")
					table.SetRowSeparator("")
					table.SetHeaderLine(false)
					table.SetBorder(false)
					table.SetTablePadding("\t") // pad with tabs
					table.SetNoWhiteSpace(true)

					for _, addon := range k8sAddons {
						templateName, _, err := util.GetTemplate(data.K8sVersionedTemplates, addon, k8sVersion)
						if err != nil {
							table.Append([]string{addon, fmt.Sprintf("%v", err)})
							continue
						}
						table.Append([]string{addon, templateName})
					}

					table.Render()

					if c.Bool("verbose") {
						replyMessage = fmt.Sprintf("%s\n%s", replyMessage, tableString.String())
						fmt.Printf(replyMessage)
						return nil
					}
					fmt.Printf("%s", tableString.String())
					return nil

				},
			},
			{
				Name:    "diffk8saddons",
				Aliases: []string{"dka"},
				Usage:   "diff 2 k8s version addons",
				Action: func(c *cli.Context) error {
					commandUsage := fmt.Sprintf("Usage: %s <k8s_version1> <k8s_version2> <rancher_version> <channel1> <channel2>", c.Command.FullName())

					if c.Args().Len() < 5 {
						return fmt.Errorf("Not enough parameters\n%s", commandUsage)
					}
					k8sVersion1 := c.Args().Get(0)
					k8sVersion2 := c.Args().Get(1)
					version := c.Args().Get(2)
					channel1 := c.Args().Get(3)
					channel2 := c.Args().Get(4)

					var dataVersion1 kdm.Data
					var dataVersion2 kdm.Data
					var err error

					dataVersion1, err = util.GetDataForChannel(version, channel1)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}
					dataVersion2, err = util.GetDataForChannel(version, channel2)
					if err != nil {
						return fmt.Errorf("Error while trying to get KDM data, error [%v]", err)
					}

					tableString := &strings.Builder{}
					table := tablewriter.NewWriter(tableString)

					table.SetHeader([]string{"Addon", k8sVersion1, k8sVersion2, "Diff?"})
					table.SetAutoWrapText(false)
					table.SetAutoFormatHeaders(true)
					table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
					table.SetAlignment(tablewriter.ALIGN_LEFT)
					table.SetCenterSeparator("")
					table.SetColumnSeparator("")
					table.SetRowSeparator("")
					table.SetHeaderLine(false)
					table.SetBorder(false)
					table.SetTablePadding("\t") // pad with tabs
					table.SetNoWhiteSpace(true)

					k8sAddons := util.GetAddonNames(dataVersion1.K8sVersionedTemplates)

					for _, addon := range k8sAddons {
						templateName1, _, err := util.GetTemplate(dataVersion1.K8sVersionedTemplates, addon, k8sVersion1)
						if err != nil {
							templateName1 = fmt.Sprintf("Error: %v", err)
						}
						templateName2, _, err := util.GetTemplate(dataVersion2.K8sVersionedTemplates, addon, k8sVersion2)
						if err != nil {
							templateName2 = fmt.Sprintf("Error: %v", err)
						}
						diff := "No"
						if templateName1 != templateName2 {
							diff = "Yes"
						}
						table.Append([]string{addon, templateName1, templateName2, diff})
					}

					table.Render()
					var replyMessage string
					if c.Bool("verbose") {
						replyMessage = fmt.Sprintf("Data for [%s] for Rancher version [%s] and channel [%s]\n", k8sVersion1, version, channel1)
						replyMessage = fmt.Sprintf("%sData for [%s] for Rancher version [%s] and channel [%s]\n", replyMessage, k8sVersion2, version, channel2)
					}

					replyMessage = fmt.Sprintf("%s%s", replyMessage, tableString.String())
					fmt.Printf(replyMessage)

					return nil

				},
			},
		},
	}
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "more verbose output",
		},
		&cli.BoolFlag{
			Name:  "diff-oneway",
			Usage: "generate diff one-way instead of default two-way",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
