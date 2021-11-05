package testing

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

func toFileName(name string) string {
	lowerName := strings.ToLower(name)
	return strings.ReplaceAll(lowerName, " ", "_")
}

// ResourceSnapshotTest takes in a test name, a resource type and a set of resources of that type. The resources are
// marshalled into yaml and written to a file under snapshots/<test_name>/<resource_type>-temp.yaml file. Creating
// the intermediary directory if necessary.
//
// The contents of the file is then compared with the contents of the equivalent file without "-temp", if this
// file doesn't exist or the contents don't match then the test fails. Otherwise it passes.
//
// If the test passes the "-temp" file is removed
//
// If the test fails it is up to the user do determine whether the changes are desired or not. If they are then
// the changes should be moved from the "-temp" file into the none temp file. (Like updating a snapshot)
//
// This function must be called from within an ginkgo IT block
func ResourceSnapshotTest(testName string, resourceType string, resources map[types.NamespacedName]interface{}) {
	resourcesData := map[string]interface{}{}
	for resName, res := range resources {
		name := fmt.Sprintf("%s/%s", resName.Namespace, resName.Name)
		resourcesData[name] = res
	}

	directory := fmt.Sprintf("snapshots/%s", toFileName(testName))
	fileName := fmt.Sprintf("%s/%s.yaml", directory, strings.ToLower(resourceType))
	tempfileName := fmt.Sprintf("%s/%s-temp.yaml", directory, strings.ToLower(resourceType))
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		By(fmt.Sprintf("Creating directory for new instance - %s", testName))
		err = os.Mkdir(directory, 0766)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to create %s directory", directory))
	}

	data, err := json.MarshalIndent(resourcesData, "", "    ")
	Expect(err).ToNot(HaveOccurred(), "Failed to marshal resources to json")
	data, err = yaml.JSONToYAML(data)
	Expect(err).ToNot(HaveOccurred(), "Failed to convert resources to yaml")
	f, err := os.Create(tempfileName)
	Expect(err).ToNot(HaveOccurred(), "Failed to create new temp file")
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.Write(data)
	Expect(err).ToNot(HaveOccurred(), "Failed to write data to buffered writer")
	err = w.Flush()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to flush buffered data to file - %s", tempfileName))
	f.Close()

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		Fail(fmt.Sprintf("File not found for %s - %s. Created file, to save remove \"-temp\"", testName, resourceType))
	}

	buffer, err := ioutil.ReadFile(fileName)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to read file - %s", fileName))
	Expect(string(data)).To(MatchYAML(string(buffer)), fmt.Sprintf("Snapshots do not match for %s - %s. Created \"-temp\" file, to save move changes into the non temp file", testName, resourceType))
	err = os.Remove(tempfileName)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed clean up temp file - %s", tempfileName))
}
