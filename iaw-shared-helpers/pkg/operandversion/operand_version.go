// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// OCO Source Materials
// 5900-AEO
//
// Copyright IBM Corp. 2021
//
// The source code for this program is not published or otherwise
// divested of its trade secrets, irrespective of what has been
// deposited with the U.S. Copyright Office.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package operandversion

import (
	"fmt"

	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/common"
)

// AvailableVersions defines the functions necessary for determining
// the appropriate reconcile version
type AvailableVersions interface {
	Versions(currentVersion string) []string
	Channels(currentVersion string) []string
	LatestForChannel(currentVersion string, channel string) string
}

// Determine the appropriate reconcile version based on the requested version/channel
// the current version and the available versions
func Determine(requestedVersion string, currentVersion string, availableVersions AvailableVersions) (string, error) {
	if common.StringSliceContains(availableVersions.Versions(currentVersion), requestedVersion) {
		return requestedVersion, nil
	}
	if common.StringSliceContains(availableVersions.Channels(currentVersion), requestedVersion) {
		return availableVersions.LatestForChannel(currentVersion, requestedVersion), nil
	}
	return "", fmt.Errorf("Failed to interpret specified version: %s", requestedVersion)
}
