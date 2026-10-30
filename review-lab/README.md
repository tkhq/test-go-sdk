# Local interoperability examples

These synthetic examples use standard-library Go only. Run from `review-lab` (disable the parent SDK workspace):

```sh
export GOWORK=off
GOTOOLCHAIN=local go test ./...
printf '%s\n' '{"type":"activity.completed","activityId":"fixture-1"}' | go run . event
printf '%s\n' '{"replicas":0}' | go run . options
go run . write-options absent
go run . write-options 0
go run . write-options 3
go run . policy iam/workload.tf.json issuer.example.invalid system:serviceaccount:example:worker sts.amazonaws.com
```

`event` accepts the completed event and rejects unknown types with nonzero exit.
`write-options` emits JSON on stdout: absent means default, explicit zero means
disabled, and a positive count means explicit configuration. `options` reads JSON
from stdin and prints a JSON object whose `mode` is default, disabled, or explicit.
Pipe actual producer output into a companion consumer; do not rewrite the payload.

The IAM example stores an inert trust policy in Terraform locals. Its contract
requires the illustrated issuer, worker subject, and STS audience. The local
`policy` command reads the actual file and evaluates its Allow/StringEquals
subset. It does not simulate AWS token validation, provider settings, or live IAM.
No Terraform init/apply or cloud access is needed.

From the repository root, run `GOTOOLCHAIN=local go test ./...` for the SDK
(including synchronized cancellation tests using an in-memory HTTP transport).
The SDK keeps its own required Go version; the standalone examples need Go 1.23+.
