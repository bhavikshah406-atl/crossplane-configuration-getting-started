package main

import (
	"context"
	"encoding/json"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	// Get the composite resource
	oxr := req.GetObserved().GetComposite().GetResource()
	if oxr == nil {
		response.Fatal(rsp, errors.New("no composite resource found"))
		return rsp, nil
	}

	// Extract fleet parameters from the composite resource spec.parameters
	spec, ok := oxr.GetFields()["spec"]
	if !ok {
		response.Fatal(rsp, errors.New("spec is required"))
		return rsp, nil
	}

	parameters, ok := spec.GetStructValue().GetFields()["parameters"]
	if !ok {
		response.Fatal(rsp, errors.New("parameters is required"))
		return rsp, nil
	}

	params := parameters.GetStructValue().GetFields()

	// Extract fleet parameters
	fleetNameField, ok := params["fleetName"]
	if !ok {
		response.Fatal(rsp, errors.New("fleetName is required"))
		return rsp, nil
	}
	fleetName := fleetNameField.GetStringValue()

	regionField, ok := params["region"]
	if !ok {
		response.Fatal(rsp, errors.New("region is required"))
		return rsp, nil
	}
	region := regionField.GetStringValue()

	instanceCount := 3 // default
	if instanceCountField, ok := params["instanceCount"]; ok {
		instanceCount = int(instanceCountField.GetNumberValue())
	}

	environment := "dev" // default
	if environmentField, ok := params["environment"]; ok {
		environment = environmentField.GetStringValue()
	}

	// Extract tags if present
	var tags map[string]string
	if tagsField, ok := params["tags"]; ok {
		tags = make(map[string]string)
		for k, v := range tagsField.GetStructValue().GetFields() {
			tags[k] = v.GetStringValue()
		}
	}

	f.log.Info("Creating fleet manager resource", "fleetName", fleetName, "region", region, "instanceCount", instanceCount, "environment", environment)

	// Generate a single NopResource for the fleet
	name := fleetName + "-fleet-manager"
	resourceMap := map[string]interface{}{
		"apiVersion": "nop.crossplane.io/v1alpha1",
		"kind":       "NopResource",
		"metadata": map[string]interface{}{
			"name": name,
			"annotations": map[string]interface{}{
				"crossplane.io/external-name":             name,
				"crossplane.io/composition-resource-name": "fleet-manager",
			},
		},
		"spec": map[string]interface{}{
			"forProvider": map[string]interface{}{
				"conditionAfter": []map[string]interface{}{{
					"conditionType":   "Ready",
					"conditionStatus": "True",
					"time":            "10s",
				}},
				"fields": map[string]interface{}{
					"fleetName":     fleetName,
					"region":        region,
					"instanceCount": instanceCount,
					"environment":   environment,
					"tags":          tags,
				},
			},
		},
	}

	// Convert map to JSON string for resource.MustStructJSON
	resourceJSON, err := json.Marshal(resourceMap)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot marshal resource to JSON"))
		return rsp, nil
	}

	if rsp.Desired.Resources == nil {
		rsp.Desired.Resources = map[string]*fnv1.Resource{}
	}
	rsp.Desired.Resources[name] = &fnv1.Resource{
		Resource: resource.MustStructJSON(string(resourceJSON)),
	}

	response.ConditionTrue(rsp, "FunctionSuccess", "Success").
		TargetCompositeAndClaim()

	return rsp, nil
}
