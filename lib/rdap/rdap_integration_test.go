// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package rdap

import (
	"context"
	"testing"
	"time"
)

// These tests make real network requests to RDAP servers.
// Run with: go test -tags=integration ./lib/rdap/

func TestIntegrationQueryIPv4(t *testing.T) {
	client := NewClient(30 * time.Second)
	ctx := context.Background()

	resp := client.QueryIP(ctx, "8.8.8.8")

	if resp.Error != "" {
		t.Fatalf("QueryIP(8.8.8.8) error: %s", resp.Error)
	}
	if resp.StartAddress == "" {
		t.Error("StartAddress should not be empty")
	}
	if resp.EndAddress == "" {
		t.Error("EndAddress should not be empty")
	}
	if resp.Name == "" {
		t.Error("Name should not be empty")
	}
	if resp.Body == "" {
		t.Error("Body should not be empty")
	}

	t.Logf("RDAP response for 8.8.8.8:\n%s", resp.Body)
}

func TestIntegrationQueryIPv6(t *testing.T) {
	client := NewClient(30 * time.Second)
	ctx := context.Background()

	resp := client.QueryIP(ctx, "2001:4860:4860::8888")

	if resp.Error != "" {
		t.Fatalf("QueryIP(2001:4860:4860::8888) error: %s", resp.Error)
	}
	if resp.StartAddress == "" {
		t.Error("StartAddress should not be empty")
	}
	if resp.Name == "" {
		t.Error("Name should not be empty")
	}
	if resp.Body == "" {
		t.Error("Body should not be empty")
	}

	t.Logf("RDAP response for 2001:4860:4860::8888:\n%s", resp.Body)
}

func TestIntegrationQueryAPNIC(t *testing.T) {
	client := NewClient(30 * time.Second)
	ctx := context.Background()

	resp := client.QueryIP(ctx, "1.1.1.1")

	if resp.Error != "" {
		t.Fatalf("QueryIP(1.1.1.1) error: %s", resp.Error)
	}
	if resp.Name == "" {
		t.Error("Name should not be empty")
	}

	t.Logf("RDAP response for 1.1.1.1:\n%s", resp.Body)
}

func TestIntegrationQueryRIPE(t *testing.T) {
	client := NewClient(30 * time.Second)
	ctx := context.Background()

	resp := client.QueryIP(ctx, "193.0.0.1")

	if resp.Error != "" {
		t.Fatalf("QueryIP(193.0.0.1) error: %s", resp.Error)
	}

	t.Logf("RDAP response for 193.0.0.1:\n%s", resp.Body)
}

func TestIntegrationHandle(t *testing.T) {
	ctx := context.Background()

	resp := Handle(ctx, "8.8.8.8")

	if resp.Error != "" {
		t.Fatalf("Handle(8.8.8.8) error: %s", resp.Error)
	}
	if resp.Query != "8.8.8.8" {
		t.Errorf("Query = %q, want %q", resp.Query, "8.8.8.8")
	}

	t.Logf("Handle response for 8.8.8.8:\n%s", resp.Body)
}
