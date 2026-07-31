package acp

import (
	"encoding/json"
	"testing"
)

func TestSessionConfigSelectOptions_UnmarshalArrayVariants(t *testing.T) {
	t.Run("ungrouped", func(t *testing.T) {
		var got SessionConfigSelectOptions
		payload := []byte(`[{"name":"fast","value":"fast"}]`)
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal ungrouped options: %v", err)
		}
		if got.Ungrouped == nil {
			t.Fatal("expected ungrouped variant to be set")
		}
		if got.Grouped != nil {
			t.Fatal("expected grouped variant to be nil")
		}
		if len(*got.Ungrouped) != 1 {
			t.Fatalf("expected one ungrouped option, got %d", len(*got.Ungrouped))
		}
		if (*got.Ungrouped)[0].Value != SessionConfigValueId("fast") {
			t.Fatalf("unexpected option value: %q", (*got.Ungrouped)[0].Value)
		}
	})

	t.Run("grouped", func(t *testing.T) {
		var got SessionConfigSelectOptions
		payload := []byte(`[{"group":"performance","name":"Performance","options":[{"name":"Balanced","value":"balanced"}]}]`)
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal grouped options: %v", err)
		}
		if got.Grouped == nil {
			t.Fatal("expected grouped variant to be set")
		}
		if got.Ungrouped != nil {
			t.Fatal("expected ungrouped variant to be nil")
		}
		if len(*got.Grouped) != 1 {
			t.Fatalf("expected one group, got %d", len(*got.Grouped))
		}
		if len((*got.Grouped)[0].Options) != 1 {
			t.Fatalf("expected one option in first group, got %d", len((*got.Grouped)[0].Options))
		}
		if (*got.Grouped)[0].Options[0].Value != SessionConfigValueId("balanced") {
			t.Fatalf("unexpected grouped option value: %q", (*got.Grouped)[0].Options[0].Value)
		}
		if _, err := json.Marshal(got); err != nil {
			t.Fatalf("marshal grouped options union: %v", err)
		}
	})
}

func TestSessionConfigOptionSelect_MetadataRoundTrip(t *testing.T) {
	modelCategory := SessionConfigOptionCategoryModel
	in := SessionConfigOption{
		Select: &SessionConfigOptionSelect{
			Type:         "select",
			Id:           SessionConfigId("reasoning_effort"),
			Name:         "Reasoning effort",
			Description:  Ptr("Controls thought depth"),
			CurrentValue: SessionConfigValueId("medium"),
			Options: SessionConfigSelectOptions{Ungrouped: &SessionConfigSelectOptionsUngrouped{
				{Name: "Low", Value: SessionConfigValueId("low")},
				{Name: "Medium", Value: SessionConfigValueId("medium")},
			}},
			Category: &modelCategory,
		},
	}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if raw["id"] != "reasoning_effort" || raw["name"] != "Reasoning effort" {
		t.Fatalf("missing identity metadata in json: %s", string(b))
	}

	var out SessionConfigOption
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if out.Select == nil || out.Select.Id != "reasoning_effort" || out.Select.Name != "Reasoning effort" {
		t.Fatalf("identity metadata lost on round-trip: %+v", out.Select)
	}
}

func TestSessionConfigOptionSelect_MetadataUnmarshalAndMarshal(t *testing.T) {
	payload := []byte(`{"type":"select","id":"model","name":"Model","category":"model","description":"Choose a model","currentValue":"gpt-4.1","options":[{"name":"GPT-4.1","value":"gpt-4.1"}]}`)

	var opt SessionConfigOption
	if err := json.Unmarshal(payload, &opt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if opt.Select == nil || opt.Select.Id != "model" || opt.Select.Name != "Model" {
		t.Fatalf("missing identity metadata: %+v", opt.Select)
	}

	b, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if raw["id"] != "model" || raw["name"] != "Model" {
		t.Fatalf("identity metadata not emitted in json: %s", string(b))
	}
}

func TestUnstableCreateElicitationRequest_ScopeRoundTrip(t *testing.T) {
	t.Run("form session scope", func(t *testing.T) {
		payload := []byte(`{"mode":"form","message":"Pick one","sessionId":"sess-1","toolCallId":"call-7","requestedSchema":{"type":"object","properties":{}}}`)
		var got UnstableCreateElicitationRequest
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Form == nil {
			t.Fatal("expected form variant to be set")
		}
		if got.Form.SessionId != SessionId("sess-1") {
			t.Fatalf("sessionId not preserved: %q", got.Form.SessionId)
		}
		if got.Form.ToolCallId == nil || *got.Form.ToolCallId != ToolCallId("call-7") {
			t.Fatalf("toolCallId not preserved: %v", got.Form.ToolCallId)
		}

		b, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatalf("unmarshal raw: %v", err)
		}
		if raw["sessionId"] != "sess-1" || raw["toolCallId"] != "call-7" {
			t.Fatalf("scope not emitted in json: %s", string(b))
		}
	})

	t.Run("url request scope", func(t *testing.T) {
		payload := []byte(`{"mode":"url","message":"Sign in","elicitationId":"el-1","url":"https://example.com","requestId":42}`)
		var got UnstableCreateElicitationRequest
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Url == nil {
			t.Fatal("expected url variant to be set")
		}
		if got.Url.SessionId != SessionId("") {
			t.Fatalf("expected empty sessionId, got %q", got.Url.SessionId)
		}
		if got.Url.RequestId == nil || got.Url.RequestId.Number == nil || *got.Url.RequestId.Number != RequestIdNumber(42) {
			t.Fatalf("requestId not preserved: %+v", got.Url.RequestId)
		}
	})
}

func TestUnstableCreateElicitationForm_PreservesRawSchemaOrder(t *testing.T) {
	// Property order is meaningful to clients (form field order); the typed
	// Properties map cannot carry it, so the raw bytes must survive.
	payload := []byte(`{"mode":"form","message":"Q","sessionId":"s","requestedSchema":{"type":"object","properties":{"zebra":{"type":"string"},"apple":{"type":"string"}},"required":["zebra"]}}`)
	var got UnstableCreateElicitationRequest
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Form == nil {
		t.Fatal("expected form variant to be set")
	}
	want := `{"type":"object","properties":{"zebra":{"type":"string"},"apple":{"type":"string"}},"required":["zebra"]}`
	if string(got.Form.RequestedSchemaRaw) != want {
		t.Fatalf("raw schema not preserved: %s", string(got.Form.RequestedSchemaRaw))
	}
	if len(got.Form.RequestedSchema.Properties) != 2 {
		t.Fatalf("typed schema not decoded: %+v", got.Form.RequestedSchema)
	}
}
