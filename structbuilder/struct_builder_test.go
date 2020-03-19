package structbuilder

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type PublishInfo struct {
	ID               int64                  `json:"@id" jq:".id"`
	Floaty           float64                `json:"floaty" jq:".floaty"`
	Booly            bool                   `json:"booly" jq:".booly"`
	StartTime        int64                  `json:"start_time" jq:".created_at"`
	EndTime          int64                  `json:"end_time" jq:".updated_at"`
	TimeSpent        int64                  `json:"time_spent" jq:".updated_at - .created_at"`
	Success          string                 `json:"success" jq:".status"`
	Bom              map[string]interface{} `json:"bom" jq:".data"`
	ArchitecturePlan string                 `json:"architecture_plan" jq:".architecture_plan.name"`
	BuildPlan        map[string]struct {
		TimeSpent     int64  `json:"time_spent" jq:".updated_at - .created_at"`
		Success       string `json:"success" jq:".status"`
		FailureReason string `json:"failure_reason" jq:".status"` // isn't a thing we have. so just use status
	} `json:"build_plan" jq:".build_results"`
}

// Load a publish info. Marshall it. Unmarshal it.
// Unmarshal known good output. Compare.
func TestRoundTrip(t *testing.T) {
	foo := PublishInfo{}
	doc := loadJSON()

	FillOut(&foo, doc)

	// quick sanity check
	if foo.ID != 123 {
		t.Errorf("Foo.ID should be 123")
	}

	jsonStr, _ := json.Marshal(foo)
	var result map[string]interface{}
	json.Unmarshal(jsonStr, &result)
	finalObj := loadFinal()

	diff := cmp.Diff(result, finalObj)
	if diff != "" {
		t.Errorf("Deep equal failed: " + diff)
	}
}

func loadJSON() map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(testJSON, &result)

	return result
}

func loadFinal() map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(roundtrip, &result)

	return result
}

// multiline string freaks vs code out
// so I put it at the bottom
var testJSON []byte = []byte(`
	{
		"id": 123,
		"floaty": 1.23,
		"booly": true,
		"created_at": 100,
		"updated_at": 200,
		"status": "success",
		"data": {
			"vxrail": "1.2.3"
		},
		"architecture_plan": {
			"name": "arch plan"
		},
		"build_results": {
			"reimage": {
				"created_at": 100,
				"updated_at": 200,
				"status": "success"
			}
		}
	}
	`)

var roundtrip []byte = []byte(`{"@id":123,"booly": true,"floaty": 1.23,"start_time":100,"end_time":200,"time_spent":100,"success":"success","bom":{"vxrail":"1.2.3"},"architecture_plan":"arch plan","build_plan":{"reimage":{"time_spent":100,"success":"success","failure_reason":"success"}}}`)
