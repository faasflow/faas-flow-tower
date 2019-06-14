package function

import (
	"encoding/json"
	"fmt"
	sdk "github.com/s8sg/faas-flow/sdk"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	NODE_CLUSTER_BORDER_COLOR = "grey"
	NODE_CLUSTER_STYLE        = "rounded"

	OPERATION_SHAPE = "rectangle"
	OPERATION_COLOR = "\"#68b8e2\""
	OPERATION_STYLE = "filled"

	EDGE_COLOR = "\"#152730\""

	CONDITION_SHAPE = "diamond"
	CONDITION_STYLE = "filled"
	CONDITION_COLOR = "\"#f9af4d\""

	FOREACH_SHAPE = "diamond"
	FOREACH_STYLE = "filled"
	FOREACH_COLOR = "\"#f9af4d\""

	DYNAMIC_END_SHAPE = "invhouse"
	DYNAMIC_END_STYLE = "filled"
	DYNAMIC_END_COLOR = "pink"

	CONDITION_CLUSTER_BORDER_COLOR = "grey"
	CONDITION_CLUSTER_STYLE        = "rounded"
)

// generateOperationKey generate a unique key for an operation
func generateOperationKey(dagId string, nodeIndex int, opsIndex int, operation *sdk.OperationExporter, operationStr string) string {
	if operation != nil {
		switch {
		case operation.IsFunction:
			operationStr = "func-" + operation.Name
		case operation.IsCallback:
			operationStr = "callback-" + operation.Name
		default:
			operationStr = "modifier"
		}
	}
	operationKey := ""
	if dagId != "0" {
		if opsIndex != 0 {
			operationKey = fmt.Sprintf("%s.%d.%d-%s", dagId, nodeIndex, opsIndex, operationStr)
		} else {
			operationKey = fmt.Sprintf("%s.%d-%s", dagId, nodeIndex, operationStr)
		}
	} else {
		operationKey = fmt.Sprintf("%d.%d-%s", nodeIndex, opsIndex, operationStr)
	}
	return operationKey
}

// generateConditionalDag generate dag element of a condition vertex
func generateConditionalDag(node *sdk.NodeExporter, dag *sdk.DagExporter, sb *strings.Builder, indent string) string {
	// Create a condition vertex
	conditionKey := generateOperationKey(dag.Id, node.Index, 0, nil, "conditions")
	sb.WriteString(fmt.Sprintf("\n%s\"%s\" [shape=%s style=%s color=%s];",
		indent, conditionKey, CONDITION_SHAPE, CONDITION_STYLE, CONDITION_COLOR))

	// Create a end operation vertex
	conditionEndKey := generateOperationKey(dag.Id, node.Index, 0, nil, "end")
	sb.WriteString(fmt.Sprintf("\n%s\"%s\" [shape=%s style=%s color=%s];",
		indent, conditionEndKey, DYNAMIC_END_SHAPE, DYNAMIC_END_STYLE, DYNAMIC_END_COLOR))

	// Create condition graph
	for condition, conditionDag := range node.ConditionalDags {
		nextOperationDag := conditionDag
		startNodeId := nextOperationDag.StartNode
		nextOperationNode := nextOperationDag.Nodes[startNodeId]

		// Find out the next node with operation (recursively)
		for nextOperationNode.SubDag != nil && !nextOperationNode.IsDynamic {
			nextOperationDag = nextOperationNode.SubDag
			startNodeId := nextOperationDag.StartNode
			nextOperationNode = nextOperationDag.Nodes[startNodeId]
		}

		operationKey := ""
		if nextOperationNode.IsDynamic {
			if nextOperationNode.IsCondition {
				operationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 0, nil, "conditions")
			}
			if nextOperationNode.IsForeach {
				operationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 0, nil, "foreach")
			}
		} else {
			operation := nextOperationNode.Operations[0]
			operationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 1, operation, "")
		}

		sb.WriteString(fmt.Sprintf("\n%s\"%s\" -> \"%s\" [label=%s color=%s];",
			indent, conditionKey, operationKey, condition, EDGE_COLOR))

		sb.WriteString(fmt.Sprintf("\n%ssubgraph cluster_%s {", indent, condition))

		sb.WriteString(fmt.Sprintf("\n%slabel=\"%s.%d-%s\";", indent+"\t", dag.Id, node.Index, condition))
		sb.WriteString(fmt.Sprintf("\n%scolor=%s;", indent+"\t", CONDITION_CLUSTER_BORDER_COLOR))
		sb.WriteString(fmt.Sprintf("\n%sstyle=%s;\n", indent+"\t", CONDITION_CLUSTER_STYLE))

		previousOperation := generateDag(conditionDag, sb, indent+"\t")

		sb.WriteString(fmt.Sprintf("\n%s}", indent))

		sb.WriteString(fmt.Sprintf("\n%s\"%s\" -> \"%s\" [color=%s];",
			indent, previousOperation, conditionEndKey, EDGE_COLOR))
	}

	return conditionEndKey
}

// generateForeachDag generate dag element of a foreach vertex
func generateForeachDag(node *sdk.NodeExporter, dag *sdk.DagExporter, sb *strings.Builder, indent string) string {
	subdag := node.SubDag

	// Create a foreach operation vertex
	foreachKey := generateOperationKey(dag.Id, node.Index, 0, nil, "foreach")
	sb.WriteString(fmt.Sprintf("\n%s\"%s\" [shape=%s style=%s color=%s];",
		indent, foreachKey, FOREACH_SHAPE, FOREACH_STYLE, FOREACH_COLOR))

	// Create a end operation vertex
	foreachEndKey := generateOperationKey(dag.Id, node.Index, 0, nil, "end")
	sb.WriteString(fmt.Sprintf("\n%s\"%s\" [shape=%s style=%s color=%s];",
		indent, foreachEndKey, DYNAMIC_END_SHAPE, DYNAMIC_END_STYLE, DYNAMIC_END_COLOR))

	// Create Foreach Graph
	{
		nextOperationDag := subdag
		startNodeId := nextOperationDag.StartNode
		nextOperationNode := nextOperationDag.Nodes[startNodeId]

		// Find out the first operation on a subdag
		for nextOperationNode.SubDag != nil && !nextOperationNode.IsDynamic {
			nextOperationDag = nextOperationNode.SubDag
			startNodeId = nextOperationDag.StartNode
			nextOperationNode = nextOperationDag.Nodes[startNodeId]
		}

		operationKey := ""
		if nextOperationNode.IsDynamic {
			if nextOperationNode.IsCondition {
				operationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 0, nil, "conditions")
			}
			if nextOperationNode.IsForeach {
				operationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 0, nil, "foreach")
			}

		} else {
			operation := nextOperationNode.Operations[0]
			operationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 1, operation, "")
		}

		sb.WriteString(fmt.Sprintf("\n%s\"%s\" -> \"%s\" [color=%s];",
			indent, foreachKey, operationKey, EDGE_COLOR))

		previousOperation := generateDag(subdag, sb, indent+"\t")

		sb.WriteString(fmt.Sprintf("\n%s\"%s\" -> \"%s\" [color=%s];",
			indent, previousOperation, foreachEndKey, EDGE_COLOR))
	}

	return foreachEndKey
}

// generateDag populate a string buffer for a dag and returns the last operation ID
func generateDag(dag *sdk.DagExporter, sb *strings.Builder, indent string) string {
	lastOperation := ""
	// gener/ate nodes
	for _, node := range dag.Nodes {

		previousOperation := ""

		if node.IsDynamic {
			// Handle dynamic node
			if node.IsCondition {
				previousOperation = generateConditionalDag(node, dag, sb, indent)
			}
			if node.IsForeach {
				previousOperation = generateForeachDag(node, dag, sb, indent)
			}
		} else {
			// Handle non dynamic node

			sb.WriteString(fmt.Sprintf("\n%ssubgraph cluster_%d {", indent, node.Index))

			nodeIndexStr := fmt.Sprintf("%d", node.Index-1)

			// Set label for node cluster
			if nodeIndexStr != node.Id {
				if dag.Id != "0" {
					sb.WriteString(fmt.Sprintf("\n%slabel=\"%s.%d-%s\";", indent+"\t", dag.Id, node.Index, node.Id))
				} else {
					sb.WriteString(fmt.Sprintf("\n%slabel=\"%d-%s\";", indent+"\t", node.Index, node.Id))
				}
			} else {
				if dag.Id != "0" {
					sb.WriteString(fmt.Sprintf("\n%slabel=\"%s-%d\";", indent+"\t", dag.Id, node.Index))
				} else {

					sb.WriteString(fmt.Sprintf("\n%slabel=\"%d\";", indent+"\t", node.Index))
				}
			}

			sb.WriteString(fmt.Sprintf("\n%scolor=%s;", indent+"\t", NODE_CLUSTER_BORDER_COLOR))
			sb.WriteString(fmt.Sprintf("\n%sstyle=%s;\n", indent+"\t", NODE_CLUSTER_STYLE))
		}

		subdag := node.SubDag
		if subdag != nil {
			previousOperation = generateDag(subdag, sb, indent+"\t")
		} else {
			for opsindex, operation := range node.Operations {
				operationKey := generateOperationKey(dag.Id, node.Index, opsindex+1, operation, "")

				sb.WriteString(fmt.Sprintf("\n%s\"%s\" [shape=%s color=%s style=%s];",
					indent+"\t", operationKey, OPERATION_SHAPE, OPERATION_COLOR, OPERATION_STYLE))

				if previousOperation != "" {
					sb.WriteString(fmt.Sprintf("\n%s\"%s\" -> \"%s\" [color=%s];",
						indent+"\t", previousOperation, operationKey, EDGE_COLOR))
				}
				previousOperation = operationKey
			}
		}

		// If noce is not dynamic close the subgraph cluster
		if !node.IsDynamic {
			sb.WriteString(fmt.Sprintf("\n%s}\n", indent))
		}

		// If node has children
		if node.Childrens != nil {
			for _, childId := range node.Childrens {

				var operation *sdk.OperationExporter

				// get child node
				child := dag.Nodes[childId]

				nextOperationNode := child
				nextOperationDag := dag

				// Find out the next node with operation (recursively)
				for nextOperationNode.SubDag != nil && !nextOperationNode.IsDynamic {
					nextOperationDag = nextOperationNode.SubDag
					nextOperationNodeId := nextOperationDag.StartNode
					nextOperationNode = nextOperationDag.Nodes[nextOperationNodeId]
				}

				childOperationKey := ""
				if nextOperationNode.IsDynamic {
					if nextOperationNode.IsCondition {
						childOperationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 0, nil, "conditions")
					}
					if nextOperationNode.IsForeach {
						childOperationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 0, nil, "foreach")
					}

				} else {
					operation = nextOperationNode.Operations[0]
					childOperationKey = generateOperationKey(nextOperationDag.Id, nextOperationNode.Index, 1, operation, "")
				}

				if previousOperation != "" {
					sb.WriteString(fmt.Sprintf("\n%s\"%s\" -> \"%s\" [color=%s];",
						indent, previousOperation, childOperationKey, EDGE_COLOR))
				}
			}
		} else {
			lastOperation = previousOperation
		}

		sb.WriteString("\n")
	}
	return lastOperation
}

// makeDotGraph make dot graph by iterating each node in greedy approach
func makeDotGraph(root *sdk.DagExporter) string {
	var sb strings.Builder

	indent := "\t"
	sb.WriteString("digraph depgraph {")
	sb.WriteString(fmt.Sprintf("\n%srankdir=TD;", indent))
	sb.WriteString(fmt.Sprintf("\n%spack=1;", indent))
	sb.WriteString(fmt.Sprintf("\n%spad=0;", indent))
	sb.WriteString(fmt.Sprintf("\n%snodesep=0;", indent))
	sb.WriteString(fmt.Sprintf("\n%sranksep=0;", indent))
	sb.WriteString(fmt.Sprintf("\n%ssplines=curved;", indent))
	sb.WriteString(fmt.Sprintf("\n%sfontname=\"Courier New\";", indent))
	sb.WriteString(fmt.Sprintf("\n%sfontcolor=\"#44413b\";", indent))

	sb.WriteString(fmt.Sprintf("\n%snode [style=filled fontname=\"Courier\" fontcolor=black]\n", indent))

	generateDag(root, &sb, indent)

	sb.WriteString("}\n")
	return sb.String()
}

// Handle a serverless request
func Handle(req []byte) string {
	values, err := url.ParseQuery(os.Getenv("Http_Query"))
	if err != nil {
		log.Fatal("No function specified")
	}

	function := values.Get("function")
	if len(function) <= 0 {
		log.Fatal("No function specified")
	}

	gateway_url := os.Getenv("gateway_url")
	if gateway_url == "" {
		gateway_url = "http://gateway:8080/"
	}

	resp, err := http.Get(gateway_url + "function/" + function + "?export-dag=true")
	if err != nil {
		log.Fatal("failed to get dag definition, ", err.Error())
	}

	defer resp.Body.Close()
	if resp.Body == nil {
		log.Fatal("failed to get dag definition, status code ", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to get dag definition, ", err.Error())
	}

	if len(bodyBytes) == 0 {
		log.Fatal("failed to get dag definition")
	}

	root := &sdk.DagExporter{}
	err = json.Unmarshal(bodyBytes, root)
	if err != nil {
		log.Fatal("failed to read dag definition, ", err.Error())
	}

	return makeDotGraph(root)
}
