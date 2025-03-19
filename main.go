package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

//==============
// ---------- 数据结构定义 ----------
type MacroEnv struct {
	Variables map[string]float64 // 存储 R 参数值
	Labels    map[string]int     // 标签行号记录（如 Label1:）
	Output    []string           // 生成的 G 代码
}

// ---------- 词法分析 ----------
var (
	reR     = regexp.MustCompile(`R(\d+)\s*=\s*(.+)`)             // R1=10
	reIf    = regexp.MustCompile(`IF\s+(.+)\s+GOTO(B|F)\s+(\w+)`) // IF R1>10 GOTOF Label1
	reLabel = regexp.MustCompile(`(\w+):`)                        // Label1:

	reSpace = regexp.MustCompile(`\s*=\s*`)
	// reX = regexp.MustCompile(`X(\d+)\s*=\s*(.+)`)         // R1=10
	// reY = regexp.MustCompile(`Y(\d+)\s*=\s*(.+)`)         // R1=10
	// reZ = regexp.MustCompile(`Z(\d+)\s*=\s*(.+)`)         // R1=10
)

// ---------- 解释器核心 ----------
func InterpretMacro(macro []string) []string {
	env := &MacroEnv{
		Variables: make(map[string]float64),
		Labels:    make(map[string]int),
		Output:    make([]string, 0),
	}

	// 预处理：扫描标签位置
	for lineNum, line := range macro {
		// fmt.Println(line, lineNum)
		if matches := reLabel.FindStringSubmatch(line); matches != nil {
			env.Labels[matches[1]] = lineNum
		}
	}
	// fmt.Println(env.Labels)
	// 逐行解释
	for lineNum := 0; lineNum < len(macro); lineNum++ {
		line := strings.TrimSpace(macro[lineNum])
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "R"):
			handleAssignment(line, env)
		case strings.HasPrefix(line, "IF"):

			lineNum = handleCondition(line, lineNum, env, macro)
		case strings.HasPrefix(line, ";"):
			continue
		case strings.HasPrefix(line, "；"):
			fmt.Println("注释符号是中文!!")
			fmt.Println(line)
			exit_1()
		default:
			expandGCode(line, env)
		}
	}
	// fmt.Println(env)
	return env.Output
}

// ---------- 新增条件判断函数 ----------
func evalCondition(condExpr string, env *MacroEnv) bool {
	// 支持比较运算符: >, <, >=, <=, ==, !=
	ops := []string{">=", "<=", "!=", "==", ">", "<"}
	for _, op := range ops {
		if strings.Contains(condExpr, op) {
			parts := strings.SplitN(condExpr, op, 2)
			left := evalExpression(strings.TrimSpace(parts[0]), env)
			right := evalExpression(strings.TrimSpace(parts[1]), env)

			switch op {
			case ">":
				return left > right
			case "<":
				return left < right
			case ">=":
				return left >= right
			case "<=":
				return left <= right
			case "==":
				return left == right
			case "!=":
				return left != right
			}
		}
	}
	panic("不支持的条件表达式: " + condExpr)
}
func handleCondition(line string, currentLine int, env *MacroEnv, macro []string) int {
	matches := reIf.FindStringSubmatch(line)
	if matches == nil {
		return currentLine
	}

	condExpr := matches[1]
	direction := matches[2] // "B" 或 "F"
	targetLabel := matches[3]

	if evalCondition(condExpr, env) {
		targetLine, ok := env.Labels[targetLabel]
		if !ok {
			fmt.Println("未定义标签: " + targetLabel)
			exit_1()
		}

		// 根据方向调整跳转行号
		if direction == "B" { // GOTOB（向后跳转）
			if targetLine >= currentLine {
				fmt.Println("GOTOB 标签必须在当前行之前")
				exit_1()
			}
			return targetLine - 1 // 跳转到标签行（-1 抵消循环中的+1）
		} else { // GOTOF（向前跳转）
			if targetLine <= currentLine {
				fmt.Println("GOTOF 标签必须在当前行之后")
				exit_1()
			}
			return targetLine - 1
		}
	}
	return currentLine
}

func expandGCode(line string, env *MacroEnv) {
	// 替换变量（如 X=R12 → X 的实际值）
	expanded := regexp.MustCompile(`R\d+`).ReplaceAllStringFunc(line, func(m string) string {
		return fmt.Sprintf("%.3f", env.Variables[m])
	})

	if matches := reLabel.FindStringSubmatch(line); matches != nil {
		return
	}
	env.Output = append(env.Output, expanded)
}

// 处理变量赋值（如 R1=5+3*R2）
func handleAssignment(line string, env *MacroEnv) {
	matches := reR.FindStringSubmatch(line)
	if matches == nil {
		return
	}

	varName := "R" + matches[1]
	expr := matches[2]
	val := evalExpression(expr, env)
	env.Variables[varName] = val
}

// ---------- 增强表达式求值函数 ----------
func evalExpression(expr string, env *MacroEnv) float64 {
	// 替换 R 参数为实际值
	expr = regexp.MustCompile(`R(\d+)`).ReplaceAllStringFunc(expr, func(m string) string {
		if val, ok := env.Variables[m]; ok {
			return fmt.Sprintf("%.6f", val)
		}
		fmt.Println("未定义变量: " + m)
		exit_1()
		return ""

	})

	// 使用 Go 的表达式解析库（需要导入 "github.com/Knetic/govaluate"）
	expression, _ := govaluate.NewEvaluableExpression(expr)
	result, _ := expression.Evaluate(nil)
	return result.(float64)
}

//==============
func exit_1() {
	fmt.Print("即将退出...")
	var a string
	fmt.Scan(&a)
	os.Exit(1)

}

//检查
func init() {
	// fmt.Println("检查环境中...")
	_, err := os.Stat("./input.MPF")
	if err == nil {
		// fmt.Println("检查环境完成...")
		return

	} else if os.IsNotExist(err) {
		fmt.Println("文件[ input.MPF ]不存在...")
		exit_1()

	} else {
		fmt.Println("其他错误...")
		exit_1()
	}
}

func R_File() []string {
	f_strings := []string{}
	f, e := os.Open("./input.MPF")
	defer f.Close()
	if e != nil {
		fmt.Println("文件打开失败...")
		exit_1()
	}
	s := bufio.NewScanner(f)
	LineNumber := 0
	for s.Scan() {
		LineNumber++
		line_txt := s.Text()
		f_strings = append(f_strings, line_txt)

	}
	if e := s.Err(); e != nil {
		fmt.Println("扫描错误...")
		exit_1()
	}
	return f_strings
}

func main() {
	os.Remove("./output.MPF")
	f, e := os.OpenFile("./output.MPF", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if e != nil {
		fmt.Println("打开文件[output.MPF]失败")
		exit_1()
	}
	defer f.Close()
	SrcCode := R_File()

	o := InterpretMacro(SrcCode)
	for i := 0; i < len(o); i++ {

		o[i] = reSpace.ReplaceAllString(o[i], "")

		fmt.Println(o[i])
		f.WriteString(o[i] + "\r\n")
	}

}
