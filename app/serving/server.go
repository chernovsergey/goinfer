package serving

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	pb "github.com/go-code/goinfer/api"
	"google.golang.org/grpc"
)

//revive:disable:exported

type FeatureName uint8

const (
	FeatureZoneID FeatureName = iota
	FeatureBannerID
	FeatureGeo
	FeatureBrowser
	FeatureOsVersion

	TotalFeatureCount

	FeatureNameSeparator  = "XX"
	FeatureValueSeparator = "X~X"
)

func (f FeatureName) fromRequest(req *pb.Request) {
	switch f {
	case FeatureZoneID:
		req.GetZoneId()
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

func (f FeatureNameString) IntName() FeatureName {
	return featureNameFromString[f]
}

type FeatureValue uint32
type Coefficient float32

type IndexKey [2]FeatureName
type IndexValue [2]FeatureValue
type ValueMap map[IndexValue]Coefficient
type CoefficientIndex map[IndexKey]ValueMap

type Yaml map[interface{}]interface{}

func RunListener(port string) *net.Listener {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Cannot listen %s", port)
	}
	log.Printf("Listening port: %s", port)
	return &lis
}

func RunServer(config Yaml) *grpc.Server {
	grpcServer := grpc.NewServer()
	inferenceServer := NewInferencer(config)
	pb.RegisterInferencerServer(grpcServer, inferenceServer)
	return grpcServer
}

type Inferencer struct {
	features map[IndexKey]bool
	coef     CoefficientIndex
}

func NewInferencer(config Yaml) *Inferencer {

	initFeatureNameFromString()

	path := config["model"].(string)
	obj := Inferencer{
		features: make(map[IndexKey]bool),
		coef:     make(CoefficientIndex),
	}
	obj.loadModel(path)
	return &obj
}

func (inf *Inferencer) loadModel(path string) {
	// Todo
	// - load model
	// - obtain features and interactions
	// - obtain coefficients
	// - save to fast access structure
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cant open model file: %s", path)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ":")

		// parse coefficient value
		coef, err := strconv.ParseFloat(tokens[2], 64)
		if err != nil {
			log.Fatalf("Cant parse coefficient %s=%s", tokens[1], tokens[2])
		}

		// parse featurename=featurevalue value
		kv := strings.Split(tokens[1], "=")
		name, value := kv[0], kv[1]

		var mapKey IndexKey
		var mapVal IndexValue
		if !strings.Contains(name, "XX") {
			// simple features e.g. zone_id=12345

			// corner case
			if value == "fit.other" {
				continue
			}

			vint, err := strconv.Atoi(value)
			if err != nil {
				log.Fatalf("Cant parse feature %q value to int", value)
			}
			mapKey = IndexKey{featureNameFromString[FeatureNameString(name)]}
			mapVal = IndexValue{FeatureValue(vint)}
		} else {
			// double interactions e.g. zone_idXXbanner_id=1234X~X5678
			nameParts := strings.Split(name, "XX")
			mapKey = IndexKey{
				featureNameFromString[FeatureNameString(nameParts[0])],
				featureNameFromString[FeatureNameString(nameParts[1])],
			}

			valueParts := strings.Split(value, "X~X")
			mapVal = IndexValue{}
			for i, part := range valueParts {

				if part == "fit.other" {
					continue
				}

				vint, err := strconv.Atoi(part)
				if err != nil {
					log.Fatalf("Cant parse feature %q value to int", value)
				}
				mapVal[i] = FeatureValue(vint)
			}
		}
		//log.Println(line)
		//log.Println(mapKey)
		//log.Println(mapVal, coef)

		inner, ok := inf.coef[mapKey]
		if !ok {
			inner = make(ValueMap)
			inner[mapVal] = Coefficient(coef)
			inf.coef[mapKey] = inner
		} else {
			inf.coef[mapKey][mapVal] = Coefficient(coef)
		}
	}

	log.Printf("Model have loaded")
	for k, v := range inf.coef {
		log.Println(k, len(v))
	}
}

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {
	return &pb.Response{Proba: 0.0, Confidence: 1.0}, nil
}
