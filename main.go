package main

import (
	"fmt"
	"log"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// mockOutOfDateCRDFromCluster represents an out-of-date CustomResourceDefinition.
// - The 'spec' properties are missing 'bar'.
// - The 'status' properties are missing 'ready'.
const mockOutOfDateCRDFromCluster = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mycustomresources.example.com-out-of-date
spec:
  group: example.com
  names:
    kind: MyCustomResource
    listKind: MyCustomResourceList
    plural: mycustomresources
    singular: mycustomresource
  scope: Namespaced
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              foo:
                type: string
          status:
            type: object
            properties:
              message:
                type: string
    served: true
    storage: true
`

// mockUpToDateCRDFromCluster represents a fully up-to-date CustomResourceDefinition.
// It contains all the fields the Go controller expects.
const mockUpToDateCRDFromCluster = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mycustomresources.example.com-up-to-date
spec:
  group: example.com
  names:
    kind: MyCustomResource
    listKind: MyCustomResourceList
    plural: mycustomresources
    singular: mycustomresource
  scope: Namespaced
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              foo:
                type: string
              bar:
                type: integer
                format: int64
          status:
            type: object
            properties:
              message:
                type: string
              ready:
                type: boolean
    served: true
    storage: true
`

// compareSchemaProperties checks if all property keys from `expected` exist in `actual`.
func compareSchemaProperties(actual, expected map[string]apiextensionsv1.JSONSchemaProps) bool {
	allPropertiesFound := true
	for key := range expected {
		if _, ok := actual[key]; !ok {
			fmt.Printf("  FAIL: Expected property '%s' not found in CRD schema.\n", key)
			allPropertiesFound = false
		} else {
			fmt.Printf("  OK: Property '%s' found in CRD schema.\n", key)
		}
	}
	return allPropertiesFound
}

// checkCRD runs the validation logic against a given CRD YAML string.
func checkCRD(crdName, crdYaml string) {
	log.Printf("===== Running Check for: %s =====\n", crdName)

	// 1. Unmarshal the raw CRD YAML into an unstructured object.
	var unstructuredObj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(crdYaml), &unstructuredObj.Object)
	if err != nil {
		log.Fatalf("Failed to unmarshal mock CRD yaml for %s: %v", crdName, err)
	}

	// 2. Convert the unstructured object into a strongly-typed CRD struct.
	var crd apiextensionsv1.CustomResourceDefinition
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, &crd)
	if err != nil {
		log.Fatalf("Failed to convert unstructured object for %s: %v", crdName, err)
	}
	log.Println("Successfully converted unstructured object to typed CRD.")

	// 3. Define the "expected" or "up-to-date" schema properties that the controller requires.
	expectedSpecProperties := map[string]apiextensionsv1.JSONSchemaProps{
		"foo": {Type: "string"},
		"bar": {Type: "integer", Format: "int64"},
	}
	expectedStatusProperties := map[string]apiextensionsv1.JSONSchemaProps{
		"message": {Type: "string"},
		"ready":   {Type: "boolean"},
	}

	// 4. Find the v1 schema and its properties from the parsed CRD.
	var v1schema *apiextensionsv1.CustomResourceValidation
	for _, version := range crd.Spec.Versions {
		if version.Name == "v1" {
			v1schema = version.Schema
			break
		}
	}

	if v1schema == nil || v1schema.OpenAPIV3Schema == nil {
		log.Printf("Could not find v1 schema in the CRD: %s", crdName)
		return
	}

	actualCRDSchema := v1schema.OpenAPIV3Schema

	// 5. Compare the properties to see if the CRD from the cluster is "up-to-date".
	log.Println("\n--- Checking if CRD Spec schema is up-to-date ---")
	specSchema, ok := actualCRDSchema.Properties["spec"]
	if !ok {
		log.Printf("CRD schema for %s is missing 'spec' property.", crdName)
		return
	}
	specUpToDate := compareSchemaProperties(specSchema.Properties, expectedSpecProperties)

	log.Println("\n--- Checking if CRD Status schema is up-to-date ---")
	statusSchema, ok := actualCRDSchema.Properties["status"]
	if !ok {
		log.Printf("CRD schema for %s is missing 'status' property.", crdName)
		return
	}
	statusUpToDate := compareSchemaProperties(statusSchema.Properties, expectedStatusProperties)

	fmt.Println("\n--- Check complete ---")
	if specUpToDate && statusUpToDate {
		log.Println("Conclusion: The installed CustomResourceDefinition schema appears to be up-to-date.")
	} else {
		log.Println("Conclusion: The installed CustomResourceDefinition schema is out-of-date.")
		if !specUpToDate {
			log.Println("-> 'spec' properties are missing.")
		}
		if !statusUpToDate {
			log.Println("-> 'status' properties are missing.")
		}
	}
	fmt.Println() // Add a blank line for better separation
}

func main() {
	log.Println("Starting mock controller to check CRD definitions...")

	// Run the check for the out-of-date CRD
	checkCRD("Out-of-Date CRD", mockOutOfDateCRDFromCluster)

	// Run the check for the up-to-date CRD
	checkCRD("Up-to-Date CRD", mockUpToDateCRDFromCluster)
}

