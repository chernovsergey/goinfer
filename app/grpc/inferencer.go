package serving

import (
	"context"
	"time"

	pb "github.com/go-code/goinfer/api"
	"github.com/go-code/goinfer/app/metrics"
)

// Inferencer is simple implementation of grpc
// InferencerService interface described in protobuf file.
//
// In general, it is container for variables, values
// and coefficients of trained model
//
// variables object is used for fetching specific fields from
// grpc request, converting them to appropriate type
//
// values stores enumerated feature values values from serialized model,
// and used for fast coefficient access for that feature
//
// coef stores coefficients of trained model
type Inferencer struct {
	variables VariableSet
	values    KVstore
	coef      CoeffStore
}

// NewInferencer produces the instance of of server
func NewInferencer(config Yaml) *Inferencer {
	initFeatureNameFromString()
	obj := Inferencer{}
	obj.loadModel(config)
	return &obj
}

// PredictProba is the main function of this project.
// It predicts probability of outcome given input request
//
// Currently this function is supposed to compute probabilities
// for logistic regression using formula
//	 p := sigmoid ( sum of cofficients )
// TODO: But it can be easily transformed to generalized response
// predictor for any linear model
func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {

	now := time.Now()
	defer func() {
		metrics.ProbabilityLatency("predict_proba", time.Since(now).Seconds())
	}()

	var score float64
	for variable := range inf.variables {
		value, err := variable.makeValue(req, &inf.values)
		if err != nil {
			return &pb.Response{}, err
		}
		coef := inf.coef[variable][value]
		score += coef
	}
	return &pb.Response{Proba: Sigmoid(score), Confidence: 1.0}, nil
}
