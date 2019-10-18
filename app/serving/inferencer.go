package serving

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

//revive:disable:exported

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
				x:    ltoken,
				y:    rtoken,
			}
		} else {
			ftype := featureNameFromString[FeatureNameString(fname)]
			variable = Variable{size: 1, x: ftype}
			features[variable] = true

			token, _ := valuestore.Set(ftype, fval)
			value = Value{
				size: 1,
				x:    token,
			}
		}

		inner, ok := coefstore[variable]
		if !ok {
			inner = make(ValueStore)
			inner[value] = c
			coefstore[variable] = inner
		} else {
			coefstore[variable][value] = c
		}
	}

	return valuestore, &features, &coefstore, nil
}
