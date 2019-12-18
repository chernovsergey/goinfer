package serving

//revive:disable:exported

import (
	"fmt"
	"strconv"

	pb "github.com/go-code/goinfer/api"
)

type FeatureName uint8

const (
	FeatureZoneID FeatureName = iota
	FeatureBannerID
	FeatureGeo
	FeatureBrowser
	FeatureOsVersion

	TotalFeatureCount     = FeatureOsVersion + 1
	FeatureNameSeparator  = "XX"
	FeatureValueSeparator = "X~X"
)

func (f FeatureName) fromRequest(req *pb.Request) (string, error) {
	switch f {
	case FeatureZoneID:
		return strconv.Itoa(int(req.GetZoneId())), nil
	case FeatureBannerID:
		return strconv.Itoa(int(req.GetBannerId())), nil
	case FeatureGeo:
		return req.GetGeo(), nil
	case FeatureBrowser:
		return strconv.Itoa(int(req.GetBrowser())), nil
	case FeatureOsVersion:
		return req.GetOsVersion(), nil
	default:
		return "", fmt.Errorf("unknown request feature %v", f.StringName())
	}

}

type FeatureNameString string

var featureNameToString = map[FeatureName]FeatureNameString{
	FeatureZoneID:    "zone_id",
	FeatureBannerID:  "banner_id",
	FeatureGeo:       "geo",
	FeatureBrowser:   "browser",
	FeatureOsVersion: "os_version",
}

var featureNameFromString = make(
	map[FeatureNameString]FeatureName, TotalFeatureCount,
)

func initFeatureNameFromString() {
	for f := FeatureZoneID; f < TotalFeatureCount; f++ {
		featureNameFromString[featureNameToString[f]] = f
	}
}

func (f FeatureName) StringName() FeatureNameString {
	return featureNameToString[f]
}

type Yaml map[interface{}]interface{}

// Variable is an abstraction for handling
// model factors as interaction of one or
// two variables.
// Is used for fetching and packing fields
// of grpc request
type Variable struct {
	size uint8
	x, y FeatureName
}

func (v Variable) makeValue(req *pb.Request, kv *KVstore) (Value, error) {
	switch v.size {
	case 1:
		val, _ := v.x.fromRequest(req)
		res, _ := kv.Get(v.x, val)
		return Value{size: 1, x: res}, nil
	case 2:
		val1, _ := v.x.fromRequest(req)
		val2, _ := v.y.fromRequest(req)
		res1, _ := kv.Get(v.x, val1)
		res2, _ := kv.Get(v.y, val2)
		return Value{size: 2, x: res1, y: res2}, nil
	default:
		return Value{}, fmt.Errorf("Nothing to return")
	}
}

func (v Variable) String() string {
	switch v.size {
	case 0:
		return fmt.Sprintf("{}")
	case 1:
		return fmt.Sprintf("{%v}",
			featureNameToString[v.x],
		)
	case 2:
		return fmt.Sprintf("{%v, %v}",
			featureNameToString[v.x],
			featureNameToString[v.y],
		)
	default:
		return fmt.Sprintf("{}")
	}
}

// Value is an abstraction for handling
// model factor values, which might contain
// one or two values
type Value struct {
	size uint8
	x, y uint32
}

type VariableSet map[Variable]bool
type ValueStore map[Value]float64
type CoeffStore map[Variable]ValueStore
