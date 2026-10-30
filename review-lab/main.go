package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Event struct {
	Type       string `json:"type"`
	ActivityID string `json:"activityId"`
}

func ConsumeEvent(data []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return event, err
	}
	if event.Type != "activity.completed" || event.ActivityID == "" {
		return event, fmt.Errorf("unsupported event")
	}
	return event, nil
}

func ConsumeOptions(data []byte) (string, error) {
	var options struct {
		Replicas *int `json:"replicas"`
	}
	if err := json.Unmarshal(data, &options); err != nil {
		return "", err
	}
	if options.Replicas == nil {
		return "default", nil
	}
	if *options.Replicas < 0 {
		return "", fmt.Errorf("negative replicas")
	}
	if *options.Replicas == 0 {
		return "disabled", nil
	}
	return "explicit", nil
}

// EvaluatePolicy interprets only this fixture's Allow/StringEquals subset.
// The Terraform file has locals only: no resources or provider operations.
func EvaluatePolicy(path, issuer, subject, audience string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var document struct {
		Locals struct {
			Policy struct {
				Statement []struct {
					Effect    string
					Action    string
					Principal struct{ Federated string }
					Condition map[string]map[string]string
				}
			} `json:"workload_policy"`
		}
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return false, err
	}
	claims := map[string]string{issuer + ":sub": subject, issuer + ":aud": audience}
	for _, statement := range document.Locals.Policy.Statement {
		if statement.Effect != "Allow" || statement.Action != "sts:AssumeRoleWithWebIdentity" || statement.Principal.Federated != "arn:aws:iam::000000000000:oidc-provider/"+issuer {
			continue
		}
		matches := true
		for operator, conditions := range statement.Condition {
			if operator != "StringEquals" {
				return false, fmt.Errorf("unsupported operator %q", operator)
			}
			for key, value := range conditions {
				if actual, ok := claims[key]; !ok || actual != value {
					matches = false
				}
			}
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected event, options, write-options, or policy")
	}
	switch args[0] {
	case "event", "options":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if args[0] == "event" {
			event, err := ConsumeEvent(data)
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(event)
		}
		mode, err := ConsumeOptions(data)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"mode": mode})
	case "write-options":
		if len(args) != 2 {
			return fmt.Errorf("expected absent or nonnegative replica count")
		}
		var replicas *int
		if args[1] != "absent" {
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				return fmt.Errorf("invalid replicas")
			}
			replicas = &n
		}
		data, err := WriteOptions(replicas)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	case "policy":
		if len(args) != 5 {
			return fmt.Errorf("expected policy path, issuer, subject, audience")
		}
		allowed, err := EvaluatePolicy(args[1], args[2], args[3], args[4])
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]bool{"allowed": allowed})
	default:
		return fmt.Errorf("unknown command")
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
