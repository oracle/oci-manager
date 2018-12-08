/*
Copyright 2018 Oracle and/or its affiliates. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"strings"
)

// StrPtrOrNil returns a pointer to the provided string
func StrPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// IsOcid returns true if the input string matches an oci id format
func IsOcid(value string) bool {
	if value == "" {
		return false
	}

	if strings.HasPrefix(value, "ocid") && len(value) > 64 {
		return true
	}
	return false
}

// Display returns the object name as display name unless the display name is set
func Display(objName string, displayName string) *string {
	display := objName
	if displayName != "" {
		display = displayName
	}
	return &display
}
