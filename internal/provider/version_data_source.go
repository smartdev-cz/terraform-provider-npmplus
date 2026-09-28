// Copyright (c) Sander Jochems
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/smartdev-cz/terraform-provider-npmplus/internal/provider/models"
)

var _ datasource.DataSource = &VersionDataSource{}

func NewVersionDataSource() datasource.DataSource {
	return &VersionDataSource{}
}

type VersionDataSource struct {
	apiURL     string
	httpClient *http.Client
}

func (d *VersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

func (d *VersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Meta --- This data source can be used to get the current version of NPMplus.",
		Attributes: map[string]schema.Attribute{
			"major": schema.Int64Attribute{
				MarkdownDescription: "The major version number.",
				Computed:            true,
			},
			"minor": schema.Int64Attribute{
				MarkdownDescription: "The minor version number.",
				Computed:            true,
			},
			"revision": schema.Int64Attribute{
				MarkdownDescription: "The revision version number.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The full version.",
				Computed:            true,
			},
		},
	}
}

func (d *VersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if data := dataSourceConfigure(ctx, req, resp); data != nil {
		d.apiURL = data.APIURL
		d.httpClient = data.HTTPClient
	}
}

func (d *VersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *models.Version

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, d.apiURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read version, got error: %s", err))
		return
	}
	response, err := d.httpClient.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read version, got error: %s", err))
		return
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read version, got HTTP status: %s", response.Status))
		return
	}

	var health struct {
		Version string `json:"version"`
	}
	body, err := io.ReadAll(response.Body)
	if err == nil {
		err = json.Unmarshal(body, &health)
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read version, got error: %s", err))
		return
	}

	data.WriteString(ctx, health.Version, &resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
