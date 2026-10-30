package main

import "testing"

func TestEvent(t *testing.T) {
	if _, err := ConsumeEvent([]byte(`{"type":"activity.completed","activityId":"fixture-1"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := ConsumeEvent([]byte(`{"type":"unknown","activityId":"fixture-1"}`)); err == nil {
		t.Fatal("accepted unknown event")
	}
}

func TestOptions(t *testing.T) {
	for _, tc := range []struct {
		input string
		mode  string
	}{{`{}`, "default"}, {`{"replicas":0}`, "disabled"}, {`{"replicas":3}`, "explicit"}} {
		if mode, err := ConsumeOptions([]byte(tc.input)); err != nil || mode != tc.mode {
			t.Fatalf("%s: %s %v", tc.input, mode, err)
		}
	}
	n := 3
	for _, tc := range []struct {
		input *int
		want  string
	}{{nil, `{}`}, {&n, `{"replicas":3}`}} {
		data, err := WriteOptions(tc.input)
		if err != nil || string(data) != tc.want {
			t.Fatalf("%s %v", data, err)
		}
	}
}

func TestPolicy(t *testing.T) {
	for _, tc := range []struct {
		issuer, subject string
		want            bool
	}{
		{"issuer.example.invalid", "system:serviceaccount:example:worker", true},
		{"issuer.example.invalid", "system:serviceaccount:example:other", false},
		{"other.example.invalid", "system:serviceaccount:example:worker", false},
	} {
		got, err := EvaluatePolicy("iam/workload.tf.json", tc.issuer, tc.subject, "sts.amazonaws.com")
		if err != nil || got != tc.want {
			t.Fatalf("%+v: %v %v", tc, got, err)
		}
	}
}
