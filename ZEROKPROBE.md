# ZerokProbe CRD Documentation

The `ZerokProbe` CRD allows users to define probes for filtering traces based on specific criteria. These probes can be used to filter traces collected in the Kubernetes cluster. These target services need to be instrumented using the OpenTelemetry operator. You can refer to this doc <placeholder> for steps on how to instrument your services.

## CRD Structure

Below is the structure of the `ZerokProbe` CRD with an explanation for each field:

- `apiVersion`: Specifies the API version of the CRD. Currently, the only supported version is `operator.zerok.ai/v1alpha1`.
- `kind`: For this CRD, it is `ZerokProbe`.
- `metadata`: 
  - `name`: The name of the probe.

### Spec Fields

- `enabled`: Determines if the probe is active (`true`) or inactive (`false`).
- `title`: A description for the probe.

### Workloads

Set of rules applying to a specific span from a specific service. Please note that this service name is the name captured by OpenTelemetry and is different from the kubernetes service name.

In the example, the rules are applied to spans from the `orders` service. The prefix `OTEL` means that the probe should only be applied to data collected by OpenTelemetry agents. Currently, this is the only type supported.  We plan to add more types in the future.

```yaml
workloads:
  "OTEL/orders":
    rule:
      type: "rule_group" 
      condition: "AND"
      rules:
```

If there are multiple workloads for different services in a probe, the rules for the specific service are only applied to a span generated from that service. For that specific span, all other workloads will be ignored.  

### Filter

Defines the filtering criteria for a particular trace. If any of the spans in the trace match a workload, the trace is considered to have satisfied the workload. Only traces that satisfy the filter condition will be exported to the OpenTelemetry collector.

- `filter`: 
    - `type`: Currently, the only supported type is `workload`.
    - `condition`: The logical condition to apply to the filter keys (`AND`/`OR`).
    - `workload_keys`: An array of keys to include in the workload filtering.  
    - `filters`: An array of more filters. The specified condition will be applied to these filters and workload keys together.

### Rules

- `workloads`:
    - `rule`: Defines a rule or group of rules for filtering.
        - `type`: The type of rule, e.g., `rule_group`.
        - `condition`: The logical condition to apply (`AND`/`OR`).
        - `rules`: An array of individual rules.
            - `type`: The type can be an individual rule (`rule`) or another rule group. Below are the fields for a rule.
            - `id`: This is the key to which the rule should be applied. Please note that the key should match exactly with the key in the trace captured by the OpenTelemetry agent.
            - `datatype`: The data type of the rule (e.g., `integer`). Please check the data types and operators section for supported data types.
            - `operator`: The operator to apply (e.g., `greater_than`, `less_than`).  Please check the data types and operators section for supported operators.
            - `value`: The value to compare against.


### Value Specification Guidelines

- **Value Format**: Regardless of the data type being evaluated, always specify the value as a string. The value will be converted to the corresponding data type before performing any evaluations.
- **Using `in` and `not_in` Operators**: Provide a comma-separated list within a single string (e.g., "2,3,4,5").
- **Using `between` and `not_between` Operators**: Specify two values separated by a comma (e.g., "5,7") within a single string. The evaluation considers both the start and end values as part of the range.


### Group By

- `group_by`: Specifies how to group the filtered traces.
    - `hash`: A unique identifier for the group.
    - `title`: A descriptive title for the group.
    - `workload_key`: The key used to identify the workload.

### Rate Limit

- `rate_limit`: Defines the rate limiting criteria.
    - `bucket_max_size`: The maximum size of the bucket for rate limiting.
    - `bucket_refill_size`: The number of tokens added to the bucket on each tick.
    - `tick_duration`: The duration between each tick when tokens are added.

## Example

Below is an example of a `ZerokProbe` CRD that filters for 4xx HTTP status codes:

```yaml
apiVersion: operator.zerok.ai/v1alpha1
kind: ZerokProbe
metadata:
  name: 4xx.error
spec:
  enabled: true
  title: "4xx Error"
  filter:
    type: "workload"
    condition: "AND"
    workload_keys:
      - "service_name"
  group_by:
    - hash: "attributes.\"service.name\""
      title: "attributes.\"service.name\""
      workload_key: "service_name"
    - hash: "attributes.\"http.status_code\""
      title: "attributes.\"http.status_code\""
      workload_key: "service_name"
  rate_limit:
    - bucket_max_size: 100
      bucket_refill_size: 10
      tick_duration: "1s"
  workloads:
    "OTEL/service_name":
      rule:
        type: "rule_group"
        condition: "AND"
        rules:
          - type: "rule"
            id: "attributes.\"http.status_code\""
            datatype: "integer"
            operator: "greater_than"
            value: "399"
          - type: "rule"
            id: "attributes.\"http.status_code\""
            datatype: "integer"
            operator: "less_than"
            value: "500"
```

# Supported Data Types and Operators

This section outlines the supported data types and the Operators applicable to each for condition evaluation.

## Data Types

- **string**
- **integer**
- **float**
- **bool**

## Operators by Data Type

### Float

- `exists`
- `not_exists`
- `less_than`
- `less_than_equal`
- `greater_than`
- `greater_than_equal`
- `equal`
- `not_equal`
- `between`
- `not_between`
- `in`
- `not_in`

### Bool

- `exists`
- `not_exists`

### Integer

- `exists`
- `not_exists`
- `less_than`
- `less_than_equal`
- `greater_than`
- `greater_than_equal`
- `equal`
- `not_equal`
- `between`
- `not_between`
- `in`
- `not_in`

### String

- `exists`
- `not_exists`
- `matches`
- `does_not_match`
- `equal`
- `not_equal`
- `contains`
- `does_not_contain`
- `in`
- `not_in`
- `begins_with`
- `does_not_begin_with`
- `ends_with`
- `does_not_end_with`