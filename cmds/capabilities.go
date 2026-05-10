package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

type capabilityProbe struct {
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}

type writePolicyProbe struct {
	DefaultMode              string `json:"default_mode"`
	DangerousRequiresConfirm bool   `json:"dangerous_requires_confirm"`
	MaxResults               int    `json:"max_results"`
}

type capabilitiesData struct {
	API struct {
		DatascriptQuery capabilityProbe `json:"datascript_query"`
		DSLQuery        capabilityProbe `json:"dsl_query"`
		Search          capabilityProbe `json:"search"`
		TagsList        capabilityProbe `json:"tags_list"`
		StateStore      capabilityProbe `json:"state_store"`
		AppInfo         capabilityProbe `json:"app_info"`
		TagSearch       capabilityProbe `json:"tag_search"`
		PropertyList    capabilityProbe `json:"property_list"`
		PropertyGet     capabilityProbe `json:"property_get"`
	} `json:"api"`
	Graph struct {
		DBGraph               capabilityProbe `json:"db_graph"`
		TagsFieldAvailable    capabilityProbe `json:"tags_field_available"`
		RefsFieldAvailable    capabilityProbe `json:"refs_field_available"`
		CurrentGraphReachable capabilityProbe `json:"current_graph_reachable"`
	} `json:"graph"`
	WritePolicy writePolicyProbe `json:"write_policy"`
	Hints       []string         `json:"hints,omitempty"`
}

type capabilitiesResult struct {
	GeneratedAt    string               `json:"generated_at"`
	ConnectionInfo ClientConnectionInfo `json:"connection_info"`
	Data           capabilitiesData     `json:"data"`
}

func CapabilitiesCmd() *redant.Command {
	return &redant.Command{
		Use:   "capabilities",
		Short: "LLM capability probing",
		Children: []*redant.Command{
			{
				Use:   "get",
				Short: "Probe runtime capabilities for LLM/MCP orchestration",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*llmEnvelope, error) {
					start := time.Now()
					res := probeCapabilities(ctx)
					return envelopeSuccess(
						start,
						res,
						withCapabilityUsed(
							"logseq.DB.datascriptQuery",
							"logseq.DB.q",
							"logseq.App.search",
							"logseq.Editor.getAllPages",
							"logseq.App.getCurrentGraph",
						),
						withHints(res.Data.Hints...),
					), nil
				}),
			},
		},
	}
}

func probeCapabilities(ctx context.Context) *capabilitiesResult {
	out := &capabilitiesResult{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		ConnectionInfo: CurrentClientConnectionInfo(),
	}

	client := NewClient()

	// API: Datascript query
	if _, err := datalogCount(ctx, client, "[:find (count ?p) :where [?p :block/name ?name]]"); err != nil {
		out.Data.API.DatascriptQuery = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else {
		out.Data.API.DatascriptQuery = capabilityProbe{Supported: true}
	}

	// API: DSL query (native)
	_, dslErr := client.DSLQuery(ctx, "{:query [:find ?name :where [?p :block/name ?name]]}")
	if dslErr != nil {
		out.Data.API.DSLQuery = capabilityProbe{Supported: false, Reason: shortErr(dslErr)}
	} else {
		out.Data.API.DSLQuery = capabilityProbe{Supported: true}
	}

	// API: Search
	if _, err := client.Search(ctx, "logseq"); err != nil {
		out.Data.API.Search = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else {
		out.Data.API.Search = capabilityProbe{Supported: true}
	}

	// API: Tags list
	if _, err := client.GetAllTags(ctx); err != nil {
		out.Data.API.TagsList = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else {
		out.Data.API.TagsList = capabilityProbe{Supported: true}
	}

	// API: State store (probe raw method availability to avoid decode-shape false negatives)
	if _, err := client.CallAPI(ctx, "logseq.App.getStateFromStore", "ui/theme"); err != nil {
		out.Data.API.StateStore = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else {
		out.Data.API.StateStore = capabilityProbe{Supported: true}
	}

	// API: app info (probe raw method, not GetInfo fallback)
	out.Data.API.AppInfo = probeMethodAvailability(ctx, client, "logseq.App.getInfo")

	// API: optional capabilities frequently affected by graph mode/version
	out.Data.API.TagSearch = probeMethodAvailability(ctx, client, "logseq.Editor.getTagsByName", "golang")
	out.Data.API.PropertyList = probeMethodAvailability(ctx, client, "logseq.Editor.getAllProperties")
	out.Data.API.PropertyGet = probeMethodAvailability(ctx, client, "logseq.Editor.getProperty", "public")

	// Graph: basic connectivity
	if _, err := client.GetCurrentGraph(ctx); err != nil {
		out.Data.Graph.CurrentGraphReachable = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else {
		out.Data.Graph.CurrentGraphReachable = capabilityProbe{Supported: true}
	}

	// Graph: DB graph mode
	if ok, err := client.CheckCurrentIsDBGraph(ctx); err != nil {
		out.Data.Graph.DBGraph = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else {
		if ok {
			out.Data.Graph.DBGraph = capabilityProbe{Supported: true}
		} else {
			out.Data.Graph.DBGraph = capabilityProbe{Supported: false, Reason: "current graph is not DB graph"}
		}
	}

	// Graph: tags field availability
	if c, err := datalogCount(ctx, client, "[:find (count ?b) :where [?b :block/tags ?t]]"); err != nil {
		out.Data.Graph.TagsFieldAvailable = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else if c > 0 {
		out.Data.Graph.TagsFieldAvailable = capabilityProbe{Supported: true}
	} else {
		out.Data.Graph.TagsFieldAvailable = capabilityProbe{Supported: false, Reason: "no rows from :block/tags"}
	}

	// Graph: refs field availability
	if c, err := datalogCount(ctx, client, "[:find (count ?b) :where [?b :block/refs ?r]]"); err != nil {
		out.Data.Graph.RefsFieldAvailable = capabilityProbe{Supported: false, Reason: shortErr(err)}
	} else if c > 0 {
		out.Data.Graph.RefsFieldAvailable = capabilityProbe{Supported: true}
	} else {
		out.Data.Graph.RefsFieldAvailable = capabilityProbe{Supported: false, Reason: "no rows from :block/refs"}
	}

	out.Data.WritePolicy = probeWritePolicy()
	out.Data.Hints = buildCapabilityHints(out)
	return out
}

func probeWritePolicy() writePolicyProbe {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("LOGSEQ_LLM_WRITE_MODE")))
	switch mode {
	case "", "read-only", "confirm", "direct":
		if mode == "" {
			mode = "read-only"
		}
	default:
		mode = "read-only"
	}

	requireConfirm := envBoolOrDefault("LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE", true)
	maxResults := envIntOrDefault("LOGSEQ_LLM_MAX_RESULTS", 200)
	if maxResults <= 0 {
		maxResults = 200
	}

	return writePolicyProbe{
		DefaultMode:              mode,
		DangerousRequiresConfirm: requireConfirm,
		MaxResults:               maxResults,
	}
}

func buildCapabilityHints(out *capabilitiesResult) []string {
	hints := make([]string, 0, 4)
	if !out.ConnectionInfo.TokenConfigured {
		hints = append(hints, "LOGSEQ_API_TOKEN 未配置，部分能力会失败")
	}
	if !out.Data.API.DSLQuery.Supported {
		hints = append(hints, "DSL 不可用时建议回退到 Datalog 或 DSL-wrapper 提取 :query")
	}
	if !out.Data.Graph.TagsFieldAvailable.Supported && out.Data.Graph.RefsFieldAvailable.Supported {
		hints = append(hints, "当前图谱 :block/tags 为空，建议基于 :block/refs 处理标签相关查询")
	}
	if !out.Data.API.Search.Supported {
		hints = append(hints, "search API 不可用时可用 datalog + 分页策略兜底")
	}
	return hints
}

func datalogCount(ctx context.Context, client *logseq.Client, query string) (int, error) {
	raw, err := client.DatascriptQuery(ctx, query)
	if err != nil {
		return 0, err
	}

	var rows [][]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return 0, fmt.Errorf("parse datalog rows: %w", err)
	}
	if len(rows) == 0 || len(rows[0]) == 0 {
		return 0, nil
	}

	return anyToInt(rows[0][0])
}

func anyToInt(v any) (int, error) {
	switch x := v.(type) {
	case float64:
		return int(x), nil
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case json.Number:
		i64, err := x.Int64()
		if err != nil {
			return 0, err
		}
		return int(i64), nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", v)
	}
}

func shortErr(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) > 180 {
		return msg[:180] + "..."
	}
	return msg
}

func probeMethodAvailability(ctx context.Context, client *logseq.Client, method string, args ...any) capabilityProbe {
	_, err := client.CallAPI(ctx, method, args...)
	if err == nil {
		return capabilityProbe{Supported: true}
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "methodnotexist") || strings.Contains(msg, "doesn't support name") {
		return capabilityProbe{Supported: false, Reason: shortErr(err)}
	}

	// Non-MethodNotExist usually means the method exists but current args/context are not ideal.
	return capabilityProbe{Supported: true, Reason: shortErr(err)}
}

func envBoolOrDefault(key string, def bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return def
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func envIntOrDefault(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}
