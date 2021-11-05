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
package common

// CombineStringStringMaps combines a set of map[string]string into a single map
// this is done such that the last map in the set gets added last, thus replacing
// any previous values
func CombineStringStringMaps(mapSet ...map[string]string) map[string]string {
	newMap := map[string]string{}
	for _, stringMap := range mapSet {
		for key, val := range stringMap {
			newMap[key] = val
		}
	}
	return newMap
}
