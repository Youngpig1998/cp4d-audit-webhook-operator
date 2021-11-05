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

// Labels is a struct used to handle resource labels
type Labels struct {
	Labels map[string]string
}

// Add adds a key value pair to the Labels map
func (l *Labels) Add(key, value string) *Labels {
	l.Labels[key] = value
	return l
}

// Remove removes a key from the Labels map
func (l *Labels) Remove(key string) *Labels {
	delete(l.Labels, key)
	return l
}
