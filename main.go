package main
import ("bufio";"encoding/csv";"encoding/json";"fmt";"os";"path/filepath";"strconv";"strings")
func main() {
	if len(os.Args) < 2 { fmt.Fprintln(os.Stderr,"Usage: data-profile <file>"); os.Exit(1) }
	path := os.Args[1]; ext := strings.ToLower(filepath.Ext(path))
	if ext == ".jsonl" || ext == ".ndjson" { profileJSONL(path); return }
	profileCSV(path)
}
func profileCSV(path string) {
	f, _ := os.Open(path); defer f.Close()
	r := csv.NewReader(f); rows, _ := r.ReadAll()
	if len(rows) < 2 { fmt.Println(`{"error":"no data"}`); return }
	cols := len(rows[0])
	fmt.Printf(`{"file":"%s","format":"csv","columns":%d,"rows":%d,"profile":[`, path, cols, len(rows)-1)
	for i := 0; i < cols; i++ {
		if i > 0 { fmt.Println(",") }
		name := rows[0][i]; nulls := 0; nums := []float64{}; strSet := map[string]int{}
		for _, row := range rows[1:] {
			if i >= len(row) || row[i] == "" { nulls++; continue }
			v := strings.TrimSpace(row[i]); strSet[v]++
			if n, err := strconv.ParseFloat(v, 64); err == nil { nums = append(nums, n) }
		}
		fmt.Printf(`{"name":"%s","nulls":%d,"unique":%d`, name, nulls, len(strSet))
		if len(nums) > 0 {
			min, max, sum := nums[0], nums[0], 0.0
			for _, n := range nums { if n < min { min = n }; if n > max { max = n }; sum += n }
			fmt.Printf(`,"type":"numeric","min":%v,"max":%v,"avg":%.2f,"count":%d`, min, max, sum/float64(len(nums)), len(nums))
		} else { fmt.Printf(`,"type":"string"`) }
		fmt.Printf(`}`)
	}
	fmt.Println("]}")
}
func profileJSONL(path string) {
	f, _ := os.Open(path); defer f.Close()
	scanner := bufio.NewScanner(f); var sample []map[string]any
	allKeys := map[string]map[string]int{}; rowCount := 0
	for scanner.Scan() {
		rowCount++; line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil { continue }
		if len(sample) < 3 { sample = append(sample, obj) }
		for k, v := range obj {
			if allKeys[k] == nil { allKeys[k] = map[string]int{} }
			allKeys[k][fmt.Sprintf("%v", v)]++
		}
	}
	fmt.Printf(`{"file":"%s","format":"jsonl","rows":%d,"keys":%d,"schema":[`, path, rowCount, len(allKeys))
	first := true
	for k, vals := range allKeys { if !first { fmt.Println(",") }; first = false; fmt.Printf(`{"key":"%s","unique":%d}`, k, len(vals)) }
	fmt.Println(`],"samples":[`)
	for i, s := range sample { if i > 0 { fmt.Println(",") }; b, _ := json.Marshal(s); fmt.Print(string(b)) }
	fmt.Println("]}")
}
