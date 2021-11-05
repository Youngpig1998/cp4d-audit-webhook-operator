package common

import "reflect"

// NamedTransformer performs a unique merge on slices of structs containing a "Name" field. This can be used
// with mergo like:
// mergo.Merge(orignal, override, mergo.WithTransformers(NamedTransformer{types: []reflect.Type{reflect.TypeOf([]corev1.EnvVar{})}}))
// the output is not sorted, but will be consistent if the original and override slices stay in the same order
type NamedTransformer struct {
	Types []reflect.Type
}

// Transformer interface implementation for arrays of named structs
func (nt NamedTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	for _, namedType := range nt.Types {
		if typ == namedType {
			return func(dst, src reflect.Value) error {
				if dst.CanSet() {
					for i := 0; i < src.Len(); i++ {
						srcEnv := src.Index(i)
						srcName := srcEnv.FieldByName("Name")
						matchFound := false
						for i := 0; i < dst.Len(); i++ {
							dstEnv := dst.Index(i)
							dstName := dstEnv.FieldByName("Name")
							if srcName.String() == dstName.String() {
								matchFound = true
								dstEnv.Set(srcEnv)
								break
							}
						}
						if !matchFound {
							dst.Set(reflect.Append(dst, srcEnv))
						}
					}
				}
				return nil
			}
		}
	}
	return nil
}
