package iflytek

import (
	"encoding/xml"
	"errors"
	"strconv"

	"github.com/echotalk/echotalk_server/internal/speech"
)

// scoreScale 讯飞英文评测总分/各项为 0-5 分制，×20 归一化到 0-100。
const scoreScale = 20.0

// xmlNode 通用 XML 节点，用于健壮遍历讯飞评测结果（结构随题型/版本有嵌套差异）。
type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Nodes   []xmlNode  `xml:",any"`
}

// attr 取指定属性值，不存在返回空串。
func (n *xmlNode) attr(name string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// parseResult 解析讯飞 ISE 结果 XML，取句级四项分与词级分，归一化到 0-100。
func parseResult(xmlBytes []byte) (speech.EvaluateResult, error) {
	var root xmlNode
	if err := xml.Unmarshal(xmlBytes, &root); err != nil {
		return speech.EvaluateResult{}, errors.New("讯飞 ISE 结果 XML 解析失败")
	}

	scoreNode := findScoreNode(&root)
	if scoreNode == nil {
		return speech.EvaluateResult{}, errors.New("讯飞 ISE 结果无评分节点(可能拒识/乱读)")
	}

	res := speech.EvaluateResult{
		Overall:   scaled(scoreNode.attr("total_score")),
		Accuracy:  scaled(scoreNode.attr("accuracy_score")),
		Fluency:   scaled(scoreNode.attr("fluency_score")),
		Integrity: scaled(scoreNode.attr("integrity_score")),
	}

	var words []speech.WordScore
	collectWords(&root, &words)
	res.Words = words
	return res, nil
}

// findScoreNode 深度优先找首个携带句级评分属性(accuracy_score + total_score)的节点。
func findScoreNode(n *xmlNode) *xmlNode {
	if n.attr("accuracy_score") != "" && n.attr("total_score") != "" {
		return n
	}
	for i := range n.Nodes {
		if got := findScoreNode(&n.Nodes[i]); got != nil {
			return got
		}
	}
	return nil
}

// collectWords 收集所有 word 节点的 content + total_score（归一化）。
func collectWords(n *xmlNode, out *[]speech.WordScore) {
	if n.XMLName.Local == "word" {
		if c := n.attr("content"); c != "" {
			*out = append(*out, speech.WordScore{Word: c, Score: scaled(n.attr("total_score"))})
		}
	}
	for i := range n.Nodes {
		collectWords(&n.Nodes[i], out)
	}
}

// scaled 把讯飞 0-5 分字符串解析并 ×20 归一化到 0-100，clamp[0,100]；无效值按 0。
func scaled(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	v *= scoreScale
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
