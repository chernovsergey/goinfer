package serving

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
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

func (f FeatureNameString) IntName() FeatureName {
	return featureNameFromString[f]
}

type FeatureValue string
type Coefficient float32
type ValueMap map[FeatureValue]Coefficient
type CoefficientIndex map[FeatureNameString]ValueMap

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
	features map[[2]string]bool
	coef     map[string]map[string]float32
}

func NewInferencer(config Yaml) *Inferencer {

	initFeatureNameFromString()

	path := config["model"].(string)
	obj := Inferencer{
		features: make(map[[2]string]bool),
		coef:     make(map[string]map[string]float32),
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
		feature := strings.Split(tokens[1], "=")
		featureName, featureValue := feature[0], feature[1]

		if strings.Contains(featureValue, "fit.other") {
			continue
		}

		if strings.Contains(featureName, "XX") {
			parts := strings.Split(featureName, "XX")
			inf.features[[2]string{parts[0], parts[1]}] = true
		} else {
			inf.features[[2]string{featureName}] = true
		}

		inner, ok := inf.coef[featureName]
		if !ok {
			inner = make(map[string]float32)
			inner[featureValue] = float32(coef)
			inf.coef[featureName] = inner
		} else {
			inf.coef[featureName][featureValue] = float32(coef)
		}
	}

	log.Printf("Model have loaded")
	for k, v := range inf.coef {
		log.Println(k, len(v))
	}
}

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {

	var score float32
	for feature := range inf.features {
		if feature[1] == "" {
			fnum := featureNameFromString[FeatureNameString(feature[0])]
			val, _ := fnum.fromRequest(req)
			coef, ok := inf.coef[feature[0]][val]
			if !ok {
				fmt.Println(feature, fnum, val)
			}
			score += coef
		} else {
			value := [2]string{}
			for i, part := range feature {
				numname := featureNameFromString[FeatureNameString(part)]
				value[i], _ = numname.fromRequest(req)
			}
			key := feature[0] + FeatureNameSeparator + feature[1]
			val := value[0] + FeatureValueSeparator + value[1]
			coef, ok := inf.coef[key][val]
			if !ok {
				//fmt.Println(feature, key, val)
			}
			score += coef
		}
	}
	prob := 1.0 / (1 + math.Exp(-1*float64(score)))
	return &pb.Response{Proba: prob, Confidence: 1.0}, nil
}
