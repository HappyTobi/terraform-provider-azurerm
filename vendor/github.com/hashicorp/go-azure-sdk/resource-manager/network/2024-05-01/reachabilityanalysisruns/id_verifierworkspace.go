package reachabilityanalysisruns

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/recaser"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

func init() {
	recaser.RegisterResourceId(&VerifierWorkspaceId{})
}

var _ resourceids.ResourceId = &VerifierWorkspaceId{}

// VerifierWorkspaceId is a struct representing the Resource ID for a Verifier Workspace
type VerifierWorkspaceId struct {
	SubscriptionId        string
	ResourceGroupName     string
	NetworkManagerName    string
	VerifierWorkspaceName string
}

// NewVerifierWorkspaceID returns a new VerifierWorkspaceId struct
func NewVerifierWorkspaceID(subscriptionId string, resourceGroupName string, networkManagerName string, verifierWorkspaceName string) VerifierWorkspaceId {
	return VerifierWorkspaceId{
		SubscriptionId:        subscriptionId,
		ResourceGroupName:     resourceGroupName,
		NetworkManagerName:    networkManagerName,
		VerifierWorkspaceName: verifierWorkspaceName,
	}
}

// ParseVerifierWorkspaceID parses 'input' into a VerifierWorkspaceId
func ParseVerifierWorkspaceID(input string) (*VerifierWorkspaceId, error) {
	parser := resourceids.NewParserFromResourceIdType(&VerifierWorkspaceId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := VerifierWorkspaceId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

// ParseVerifierWorkspaceIDInsensitively parses 'input' case-insensitively into a VerifierWorkspaceId
// note: this method should only be used for API response data and not user input
func ParseVerifierWorkspaceIDInsensitively(input string) (*VerifierWorkspaceId, error) {
	parser := resourceids.NewParserFromResourceIdType(&VerifierWorkspaceId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := VerifierWorkspaceId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

func (id *VerifierWorkspaceId) FromParseResult(input resourceids.ParseResult) error {
	var ok bool

	if id.SubscriptionId, ok = input.Parsed["subscriptionId"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "subscriptionId", input)
	}

	if id.ResourceGroupName, ok = input.Parsed["resourceGroupName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "resourceGroupName", input)
	}

	if id.NetworkManagerName, ok = input.Parsed["networkManagerName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "networkManagerName", input)
	}

	if id.VerifierWorkspaceName, ok = input.Parsed["verifierWorkspaceName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "verifierWorkspaceName", input)
	}

	return nil
}

// ValidateVerifierWorkspaceID checks that 'input' can be parsed as a Verifier Workspace ID
func ValidateVerifierWorkspaceID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseVerifierWorkspaceID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Verifier Workspace ID
func (id VerifierWorkspaceId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/networkManagers/%s/verifierWorkspaces/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.NetworkManagerName, id.VerifierWorkspaceName)
}

// Segments returns a slice of Resource ID Segments which comprise this Verifier Workspace ID
func (id VerifierWorkspaceId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("staticSubscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("staticResourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("staticProviders", "providers", "providers"),
		resourceids.ResourceProviderSegment("staticMicrosoftNetwork", "Microsoft.Network", "Microsoft.Network"),
		resourceids.StaticSegment("staticNetworkManagers", "networkManagers", "networkManagers"),
		resourceids.UserSpecifiedSegment("networkManagerName", "networkManagerName"),
		resourceids.StaticSegment("staticVerifierWorkspaces", "verifierWorkspaces", "verifierWorkspaces"),
		resourceids.UserSpecifiedSegment("verifierWorkspaceName", "verifierWorkspaceName"),
	}
}

// String returns a human-readable description of this Verifier Workspace ID
func (id VerifierWorkspaceId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Network Manager Name: %q", id.NetworkManagerName),
		fmt.Sprintf("Verifier Workspace Name: %q", id.VerifierWorkspaceName),
	}
	return fmt.Sprintf("Verifier Workspace (%s)", strings.Join(components, "\n"))
}
