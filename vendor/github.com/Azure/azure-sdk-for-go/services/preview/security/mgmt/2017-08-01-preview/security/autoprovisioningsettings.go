package security

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"context"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"net/http"
)

// AutoProvisioningSettingsClient is the API spec for Microsoft.Security (Azure Security Center) resource provider
type AutoProvisioningSettingsClient struct {
	BaseClient
}

// NewAutoProvisioningSettingsClient creates an instance of the AutoProvisioningSettingsClient client.
func NewAutoProvisioningSettingsClient(subscriptionID string, ascLocation string) AutoProvisioningSettingsClient {
	return NewAutoProvisioningSettingsClientWithBaseURI(DefaultBaseURI, subscriptionID, ascLocation)
}

// NewAutoProvisioningSettingsClientWithBaseURI creates an instance of the AutoProvisioningSettingsClient client.
func NewAutoProvisioningSettingsClientWithBaseURI(baseURI string, subscriptionID string, ascLocation string) AutoProvisioningSettingsClient {
	return AutoProvisioningSettingsClient{NewWithBaseURI(baseURI, subscriptionID, ascLocation)}
}

// Create details of a specific setting
// Parameters:
// settingName - auto provisioning setting key
// setting - auto provisioning setting key
func (client AutoProvisioningSettingsClient) Create(ctx context.Context, settingName string, setting AutoProvisioningSetting) (result AutoProvisioningSetting, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: client.SubscriptionID,
			Constraints: []validation.Constraint{{Target: "client.SubscriptionID", Name: validation.Pattern, Rule: `^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$`, Chain: nil}}}}); err != nil {
		return result, validation.NewError("security.AutoProvisioningSettingsClient", "Create", err.Error())
	}

	req, err := client.CreatePreparer(ctx, settingName, setting)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "Create", nil, "Failure preparing request")
		return
	}

	resp, err := client.CreateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "Create", resp, "Failure sending request")
		return
	}

	result, err = client.CreateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "Create", resp, "Failure responding to request")
	}

	return
}

// CreatePreparer prepares the Create request.
func (client AutoProvisioningSettingsClient) CreatePreparer(ctx context.Context, settingName string, setting AutoProvisioningSetting) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"settingName":    autorest.Encode("path", settingName),
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2017-08-01-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPut(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/providers/Microsoft.Security/autoProvisioningSettings/{settingName}", pathParameters),
		autorest.WithJSON(setting),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// CreateSender sends the Create request. The method will close the
// http.Response Body if it receives an error.
func (client AutoProvisioningSettingsClient) CreateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// CreateResponder handles the response to the Create request. The method always
// closes the http.Response Body.
func (client AutoProvisioningSettingsClient) CreateResponder(resp *http.Response) (result AutoProvisioningSetting, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Get details of a specific setting
// Parameters:
// settingName - auto provisioning setting key
func (client AutoProvisioningSettingsClient) Get(ctx context.Context, settingName string) (result AutoProvisioningSetting, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: client.SubscriptionID,
			Constraints: []validation.Constraint{{Target: "client.SubscriptionID", Name: validation.Pattern, Rule: `^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$`, Chain: nil}}}}); err != nil {
		return result, validation.NewError("security.AutoProvisioningSettingsClient", "Get", err.Error())
	}

	req, err := client.GetPreparer(ctx, settingName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client AutoProvisioningSettingsClient) GetPreparer(ctx context.Context, settingName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"settingName":    autorest.Encode("path", settingName),
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2017-08-01-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/providers/Microsoft.Security/autoProvisioningSettings/{settingName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client AutoProvisioningSettingsClient) GetSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client AutoProvisioningSettingsClient) GetResponder(resp *http.Response) (result AutoProvisioningSetting, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// List exposes the auto provisioning settings of the subscriptions
func (client AutoProvisioningSettingsClient) List(ctx context.Context) (result AutoProvisioningSettingListPage, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: client.SubscriptionID,
			Constraints: []validation.Constraint{{Target: "client.SubscriptionID", Name: validation.Pattern, Rule: `^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$`, Chain: nil}}}}); err != nil {
		return result, validation.NewError("security.AutoProvisioningSettingsClient", "List", err.Error())
	}

	result.fn = client.listNextResults
	req, err := client.ListPreparer(ctx)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "List", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListSender(req)
	if err != nil {
		result.apsl.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "List", resp, "Failure sending request")
		return
	}

	result.apsl, err = client.ListResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "List", resp, "Failure responding to request")
	}

	return
}

// ListPreparer prepares the List request.
func (client AutoProvisioningSettingsClient) ListPreparer(ctx context.Context) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2017-08-01-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/providers/Microsoft.Security/autoProvisioningSettings", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ListSender sends the List request. The method will close the
// http.Response Body if it receives an error.
func (client AutoProvisioningSettingsClient) ListSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// ListResponder handles the response to the List request. The method always
// closes the http.Response Body.
func (client AutoProvisioningSettingsClient) ListResponder(resp *http.Response) (result AutoProvisioningSettingList, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// listNextResults retrieves the next set of results, if any.
func (client AutoProvisioningSettingsClient) listNextResults(lastResults AutoProvisioningSettingList) (result AutoProvisioningSettingList, err error) {
	req, err := lastResults.autoProvisioningSettingListPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "listNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}
	resp, err := client.ListSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "listNextResults", resp, "Failure sending next results request")
	}
	result, err = client.ListResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "security.AutoProvisioningSettingsClient", "listNextResults", resp, "Failure responding to next results request")
	}
	return
}

// ListComplete enumerates all values, automatically crossing page boundaries as required.
func (client AutoProvisioningSettingsClient) ListComplete(ctx context.Context) (result AutoProvisioningSettingListIterator, err error) {
	result.page, err = client.List(ctx)
	return
}
