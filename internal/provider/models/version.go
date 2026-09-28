// Copyright (c) Sander Jochems
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sander0542/nginxproxymanager-go"
	"strconv"
	"strings"
)

type Version struct {
	Major    types.Int64  `tfsdk:"major"`
	Minor    types.Int64  `tfsdk:"minor"`
	Revision types.Int64  `tfsdk:"revision"`
	Version  types.String `tfsdk:"version"`
}

func (m *Version) Write(_ context.Context, version *nginxproxymanager.Health200ResponseVersion, _ *diag.Diagnostics) {
	m.Major = types.Int64Value(version.GetMajor())
	m.Minor = types.Int64Value(version.GetMinor())
	m.Revision = types.Int64Value(version.GetRevision())
	m.Version = types.StringValue(fmt.Sprintf("%d.%d.%d", version.GetMajor(), version.GetMinor(), version.GetRevision()))
}

func (m *Version) WriteString(_ context.Context, version string, diagnostics *diag.Diagnostics) {
	parts := strings.SplitN(strings.TrimPrefix(version, "v"), ".", 3)
	if len(parts) != 3 {
		diagnostics.AddError("Invalid version", fmt.Sprintf("Unable to parse version %q", version))
		return
	}

	values := make([]int64, 3)
	for i, part := range parts {
		part = strings.SplitN(part, "-", 2)[0]
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			diagnostics.AddError("Invalid version", fmt.Sprintf("Unable to parse version %q: %s", version, err))
			return
		}
		values[i] = value
	}

	m.Major = types.Int64Value(values[0])
	m.Minor = types.Int64Value(values[1])
	m.Revision = types.Int64Value(values[2])
	m.Version = types.StringValue(version)
}
