package serving

import (
	"testing"
)

func TestIndexing(t *testing.T) {

	kv := NewKVStore()

	feature := FeatureName(2)

	valsForInsert := []string{
		"us", "gb", "fr", "it", "in", "de", "ru", "us",
	}

	for i := 0; i < len(valsForInsert); i++ {
		kv.Set(feature, valsForInsert[i])
	}

	for i := 0; i < len(valsForInsert); i++ {
		res, _ := kv.Get(feature, valsForInsert[i])
		if res != uint32(i) {
			t.Errorf(
				"%s != %d (but %d). Dump: \n%v\n%v",
				valsForInsert[i], i, res, kv.store, kv.uniqs,
			)
		}
	}
}
