package serving

//revive:disable:exported

import "math"

func unpackArray(arr []string, vars ...*string) {
	for i, v := range arr {
		*vars[i] = v
	}
}

func typelook(vars []string, types ...*FeatureName) {
	for i, v := range vars {
		*types[i] = featureNameFromString[FeatureNameString(v)]
	}
}

func Sigmoid(x float32) float64 {
	return 1.0 / (1 + math.Exp(-1*float64(x)))
}
