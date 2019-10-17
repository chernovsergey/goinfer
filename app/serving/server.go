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

func (f FeatureNameString) IntName() FeatureName {
	return featureNameFromString[f]
}

type FeatureValue uint32
type Coefficient float32

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

// Variable is an interaction of
// several features
type Variable struct {
	size uint8
	x, y FeatureName
}

func (v Variable) makeValue(req *pb.Request, kv *KVstore) (Value, error) {
	switch v.size {
	case 1:
		val, _ := v.x.fromRequest(req)
		res, _ := kv.Get(v.x, val)
		return Value{size: 1, x: FeatureValue(res)}, nil
	case 2:
		val1, _ := v.x.fromRequest(req)
		val2, _ := v.y.fromRequest(req)
		res1, _ := kv.Get(v.x, val1)
		res2, _ := kv.Get(v.y, val2)
		return Value{size: 2, x: FeatureValue(res1),
			y: FeatureValue(res2)}, nil
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

// Value is a set of Variable
// feature values
type Value struct {
	size uint8
	x, y FeatureValue
}

type VariableSet map[Variable]bool
type ValueStore map[Value]Coefficient
type CoeffStore map[Variable]ValueStore

type Inferencer struct {
	variables VariableSet
	values    KVstore
	coef      CoeffStore
}

func NewInferencer(config Yaml) *Inferencer {

	initFeatureNameFromString()

	obj := Inferencer{}
	obj.loadModel(config)
	return &obj
}

func (inf *Inferencer) loadModel(config Yaml) {
	path := config["model"].(string)
	lines, err := scanfile(path)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	kv, vars, coef, err := parse(lines)
	if err != nil {
		log.Fatalf("Failed to parse model: %v", err)
	}
	inf.variables = *vars
	inf.values = *kv
	inf.coef = *coef

	log.Printf("Model have loaded successfully!")
	for k, v := range inf.coef {
		log.Println(k, len(v))
	}
}

func scanfile(path string) (*[]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return &[]string{}, err
	}

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, 1000)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return &lines, nil
}

func parse(lines *[]string) (*KVstore, *VariableSet, *CoeffStore, error) {
	// Format of each line is
	// <positional No. of feature>:<name>=<value>:<coefficient>
	// Positional no. of feature is useless for inference
	// so it's just ignored
	valuestore := NewKVStore()
	features := make(VariableSet)
	coefstore := make(CoeffStore)

	for _, line := range *lines {

		var no, feature, coef string
		unpackArray(strings.Split(line, ":"), &no, &feature, &coef)

		c, err := strconv.ParseFloat(coef, 64)
		if err != nil {
			log.Fatalf("Failed to parse coefficient %s", line)
		}

		var fname, fval string
		unpackArray(strings.Split(feature, "="), &fname, &fval)

		if strings.Contains(fval, "fit.other") {
			continue
		}

		variable := Variable{}
		value := Value{}
		if strings.Contains(fname, "XX") {
			var l, r string
			unpackArray(strings.Split(fname, FeatureNameSeparator), &l, &r)

			var ltype, rtype FeatureName
			typelook([]string{l, r}, &ltype, &rtype)

			variable = Variable{size: 2, x: ltype, y: rtype}
			features[variable] = true

			var lval, rval string
			unpackArray(
				strings.Split(fval, FeatureValueSeparator),
				&lval, &rval)
			ltoken, _ := valuestore.Set(ltype, lval)
			rtoken, _ := valuestore.Set(rtype, rval)

			value = Value{
				size: 2,
				x:    FeatureValue(ltoken),
				y:    FeatureValue(rtoken),
			}
		} else {
			ftype := featureNameFromString[FeatureNameString(fname)]
			variable = Variable{size: 1, x: ftype}
			features[variable] = true

			token, _ := valuestore.Set(ftype, fval)
			value = Value{
				size: 1,
				x:    FeatureValue(token),
			}
		}

		inner, ok := coefstore[variable]
		if !ok {
			inner = make(ValueStore)
			inner[value] = Coefficient(c)
			coefstore[variable] = inner
		} else {
			coefstore[variable][value] = Coefficient(c)
		}
	}

	return valuestore, &features, &coefstore, nil
}

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

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {

	var score Coefficient
	for variable := range inf.variables {
		value, err := variable.makeValue(req, &inf.values)
		if err != nil {
			return &pb.Response{}, err
		}
		coef := inf.coef[variable][value]
		score += coef
	}
	return &pb.Response{Proba: Sigmoid(float32(score)), Confidence: 1.0}, nil
}
