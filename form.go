// formatting functions

package iof

import (
  "fmt"
  "log"
  "time"
  "bytes"
  "errors"
  "regexp"
  "slices"
  "reflect"
  "runtime"
  "strings"
  "strconv"
  "unicode"
  "crypto/md5"
  "encoding/hex"
  "encoding/json"
)

// get first row (dict) from resultset
func First(data []interface{}, structure []string) []string {
  return Row(0, data, structure)
}

// get a row (dict) from resultset["result"]
func Row(index int, data []interface{}, structure []string) []string {
  if index < len(data) {
    return Structify(data[index].(map[string]any), structure)
  } else {
    return nil
  }
}

// get all rows from resultset["result"]
func Array(data []interface{}, structure []string) [][]string {
  var rows [][]string
  for index := range data {
    row := Structify(data[index].(map[string]any), structure)
    rows = append(rows, row)
  }
  return rows
}

// get raw data (list of dicts) from resultset
func Result(data map[string]interface{}) []interface{} {
  return data[RESULT].([]interface{})
}

// get an item from structured slice
func Pick(data []string, id string, structure []string) string {
  if len(data) != len(structure) || !slices.Contains(structure, id) {
    log.Fatal(INVALIDSTRUCT())
  }
  for index := range structure {
    if structure[index] == id { return Nice(data[index]) }
  }
  return ""
}

// make data row object structured slice for output
func Structify(data map[string]any, structure []string) []string {
  clone := make([]string, len(structure))
  copy(clone, structure)
  for index, id := range structure {
    if item, exists := data[id]; exists {
      clone[index] = Nice(item)
    } else {
      log.Fatal(INVALIDSTRUCT())
    }
  }
  return clone
}

// get columns width of text matrix
func Widths(data [][]string) []int {
  cols := Wide(data)
  if cols == 0 { return []int{} }
  widths := []int{}
  for col := range cols {
    width := 0
    for row := range len(data) {
      w := len(data[row][col])
      if w > width { width = w }
    }
    widths = append(widths, width)
  }
  return widths
}

// count columns of 2d matrix
func Wide(data [][]string) int {
  if len(data) == 0 { return 0 }
  return len(data[0])
}

// generate "name" jsondata
func Jname(name string) string {
  return Stringify(map[string]interface{}{"name": name})
}

// implement javascript json.stringify
func Stringify(data map[string]interface{}) string {
  jdata, err := json.Marshal(data)
  if err != nil { log.Fatal(err) }
  return string(jdata)
}

// implement javascript json.parse
func Parse(sdata string) map[string]interface{} {
  var data map[string]interface{}
  err := json.Unmarshal([]byte(sdata), &data)
  if err != nil { log.Fatal(err) }
  return data
}

// assertion map of (string) to (any)
func Any(src map[string]string) map[string]any {
  if src == nil { return nil }
  dst := make(map[string]any, len(src))
  for key, val := range src { dst[key] = val }
  return dst
}

// assertion map of (any) to (string)
func FromAny(src map[string]any) map[string]string {
  if src == nil { return nil }
  dst := make(map[string]string, len(src))
  for key, val := range src { dst[key] = String(val) }
  return dst
}

// convert data to string
func String(data any) string {
  return fmt.Sprintf("%v", data)
}

// make nil as blank str
func Nice(data any) string {
  v := String(data)
  if v == "<nil>" { return "" }
  return  v
}

// convert string to integer
func Integer(data string) int {
  num, err := strconv.Atoi(data)
  if err == nil { return num }
  return 0
}

// convert string to float
func Float64(data string) float64 {
  num, err := strconv.ParseFloat(data, 64)
  if err == nil { return num }
  return 0
}

// use with Input() for none
func None(data any) (any, error) {
  return data, nil
}

// use with Input()
// convert data to string
func Str(data any) (any, error) {
  return any(String(data)), nil
}

// use with Input()
// convert data to integer
func Int(data any) (any, error) {
  num, err := strconv.Atoi(data.(string))
  if err == nil { return num, nil }
  return 0, err
}

// use with Input()
// convert data to float
func Float(data any) (any, error) {
  num, err := strconv.ParseFloat(data.(string), 64)
  if err == nil { return num, nil }
  return 0, err
}

// use with Input()
// convert data to Date()
func Cdate(data any) (any, error) {
  s, err := Date(String(data))
  return any(s), err
}

// use with Input()
// convert data to Time()
func Ctime(data any) (any, error) {
  s, err := Time(String(data))
  return any(s), err
}

// use with Input()
// convert data to bool
func Bool(data any) (any, error) {
  switch Lower(String(data)) {
  case "y", "yes", "true", "1": return true, nil
  default: return false, nil
  }
}

// convert string to bool
// to trigger some action
func Yes(data string) bool {
  v, _ := Bool(any(data))
  return v.(bool)
}

func StrCash(num string) string { return Cash(any(num)) }

// format float number string
// with thousands separator [,]
func Cash(num any) string {
  dec := DOT
  s := strings.TrimSpace(String(num))
  d := strings.Split(s, dec)
  if len(d) > 2 { return s }
  d1, err := strconv.Atoi(d[0])
  if True(err) { return s }
  number := money(d1)
  if len(d) == 1 { return number }
  d2, err := strconv.Atoi(d[1])
  if True(err) { return s }
  return number + dec + String(d2)
}

// format integer number string
// with thousands separator [,]
func money(n int) string {
  neg := "-"
  sign := ""
  in := strconv.Itoa(n)
  if string(in[0]) == neg {
    sign = neg
    in = in[1:]
  }
  var buf bytes.Buffer
  k := 0
  for i := len(in) - 1; i >= 0; i-- {
    buf.WriteByte(in[i])
    k++
    if k%3 == 0 && i != 0 {
      buf.WriteByte(',')
    }
  }
  return sign + reverse(buf.String())
}

func reverse(s string) string {
  runes := []rune(s)
  for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
    runes[i], runes[j] = runes[j], runes[i]
  }
  return string(runes)
}

// get md5 hash of string
func MD5(text string) string {
  data := []byte(text)
  hashBytes := md5.Sum(data)
  return hex.EncodeToString(hashBytes[:])
}

// get calling function name
// set level=0 for owner func
func Callup(level int) string {
  pc, _, _, ok := runtime.Caller(level + 1)
  if !ok { return "" }
  f := runtime.FuncForPC(pc)
  if f == nil { return "" }
  fn := strings.Split(f.Name(), DOT)
  return fn[len(fn)-1]
}

// capitalize first letter of string
func Capitalize(str string) string {
  var output []rune
  isWord := true
  for _, val := range str {
    if isWord && unicode.IsLetter(val) {
      output = append(output, unicode.ToUpper(val))
      isWord = false
    } else if !unicode.IsLetter(val) && val != '_' {
      isWord = true
      output = append(output, val)
    } else {
      output = append(output, val)
    }
  }
  return string(output)
}

func Lower(str string) string {
  return strings.ToLower(str)
}

func Upper(str string) string {
  return strings.ToUpper(str)
}

// parse menu option to function id
func Optionize(option string) string {
  // need some stuffs here
  return Lower(option)
}

// parse menu option to agent func
func Agentize(option string) string {
  return Capitalize(Optionize(option))
}

// get [map] object attr as string
func Get(data map[string]string, attr string) string {
  if item, ok := data[attr]; ok { return item }
  return ""
}

// get [map] object attr as any type
func Pull(data map[string]any, attr string) any {
  if item, ok := data[attr]; ok { return item }
  return nil
}

func IsStr(value any) bool {
  _, ok := value.(string)
  return ok
}

func IsBool(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Bool
}

func IsInt(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Int
}

func IsFloat(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Float64
}

func IsMap(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Map
}

func IsArray(data any) bool {
  v := reflect.ValueOf(data)
  return (v.Kind() == reflect.Array || v.Kind() == reflect.Slice)
}

// check two vars equal
func Matched(var1 any, var2 any) bool {
  return True(var2) && var2 == var1
}

// check var as boolean
func False(data any) bool {
  return !True(data)
}

// check var as boolean
func True(data any) bool {
  if data == nil { return false }
  if IsBool(data) { return data.(bool) }
  if IsStr(data) && data == "" { return false }
  if (IsInt(data) || IsFloat(data)) && data == 0 { return false }
  if (IsMap(data) || IsArray(data)) && reflect.ValueOf(data).Len() == 0 { return false }
  return true
}

// retrieve maps in data slice where specific key matches given term
func Search(data []map[string]string, key, term string) []map[string]string {
  var results []map[string]string
  for _, m := range data {
    if val, ok := m[key]; ok && val == term {
      results = append(results, m)
    }
  }
  return results
}

// retrieve single first map of Search() function
func Find(data []map[string]string, key, term string) map[string]string {
  results := Search(data, key, term)
  if len(results) > 0 { return results[0] }
  return map[string]string{}
}

// get valid menu option index+1
func Optioned(choice string, total int) int {
  num := Integer(choice)
  if num < 1 || num > total { return 0 }
  return num
}

// get valid menu multi-options index+1
func MultiOptioned(choices []string, total int) []int {
  var options []int
  for _, choice := range choices {
    num := Optioned(choice, total)
    if num == 0 { return []int{} }
    options = append(options, num)
  }
  return options
}

// validate agent function
func Valid(funcname string) bool {
  return reflect.ValueOf(Standalone{}).MethodByName(funcname).IsValid()
}

var Agented = Valid // alias

// apply bool to db bool
func DbBool(data bool) int {
  if data { return 1 }
  return 0
}

// Convert str bool to yes/no
func StrBool(data string) string {
  if Include(any(data), []any{"<nil>", nil, "false", false, "0", 0, ""}) { return LABELS["NO"] }
  return LABELS["YES"]
}

// shorten sensitive string
func StrZip(text string) string {
  if len(text) < 5 { return THREEDOT }
  return text[0:5] + THREEDOT
}

// convert db time to date
func StrDate(time string) string {
  if False(time) || Snil(time) || len(time) < 10 { return "" }
  return time[0:10]
}

// Convert uk date to rfc3339
func Date(data string) (string, error) {
  const layout = "2/1/2006"
  t, err := time.Parse(layout, data)
  if True(err) { return "", errors.New(INVALIDINPUT()) }
  return t.Format(time.RFC3339)[0:10], nil
}

// Adjust hour:minute as rfc3339
func Time(data string) (string, error) {
  const layout = "2006-01-02 15:4"
  data = "2006-01-02 " + data
  t, err := time.Parse(layout, data)
  if True(err) { return "", errors.New(INVALIDINPUT()) }
  return t.Format(time.RFC3339)[11:16], nil
}

// split string to array
func Split(data string, separator string) []string {
  slice := strings.Split(data, separator)
  for i, val := range slice { slice[i] = strings.TrimSpace(val) }
  return slice
}

// extract string to array
func Extract(data string) []string {
  slice := Split(data, COMMA)
  for i, val := range slice { slice[i] = strings.TrimSpace(val) }
  return slice
}

// extract string to int interval
func Interval(data string) []int {
  slice := Split(data, HYPHEN)
  return []int{Integer(slice[0]), Integer(slice[len(slice)-1])}
}

// extract string to array of int
func Integers(data string) []int {
  slice := Extract(data)
  inums := make([]int, len(slice))
  for i, val := range slice { inums[i] = Integer(val) }
  return inums
}

// clone python f-string with [%v] values
func Fstr(rawstr string, agrs ...any) string {
  return fmt.Sprintf(rawstr, agrs...)
}

// clear all puncs in string
func Escape(data string) string {
  reg := regexp.MustCompile("[^\\p{L}\\p{N}]+")
  return reg.ReplaceAllString(data, "")
}

// reformat input string for names and ids
func Form(data string, format string) string {
  forms := Extract(format)
  for _, form := range forms {
    switch Lower(form) {
    case "upper", "up", "u":
      data = Upper(data)
    case "lower", "low", "l":
      data = Lower(data)
    case "clear", "clr", "c":
      data = Escape(data)
    }
  }
  return data
}

// check integer in range
func Ranged(val int, data []int) bool {
  return len(data) == 2 && val >= data[0] && val <= data[1]
}

// check integer in slice
func In(val int, data []int) bool {
  return slices.Contains(data, val)
}

// check any value in slice
func Include(val any, data []any) bool {
  return slices.Contains(data, val)
}

// convert server response interface to map
func Map(data []interface{}) []map[string]any {
  result := make([]map[string]any, len(data))
  for index, val := range data {
    if mapval, ok := val.(map[string]any); ok {
      result[index] = mapval
    } else {
      return nil
    }
  }
  return result
}

func Now() string { return time.Now().Format(time.RFC3339) }

func Update(data map[string]interface{}, key string, value any) { data[key] = value }
func Default(data map[string]interface{}, key string, value any) { if False(Pull(data, key)) { data[key] = value } }

func Type(data any) string { return fmt.Sprintf("%T", data) } // get var type as string
func Snil(data string) bool { return data == "<nil>" } // caution: check string as nil
func Nil(data any) bool { return data == nil } // check var as nil

// extract str to int range
func Ints(data string) ([]int, bool) {
  if strings.Contains(data, HYPHEN) {
    return Interval(data), true
  } else if strings.Contains(data, COMMA) {
    return Integers(data), false
  }
  return []int{Integer(data)}, false
}

// extract str to str range
func Strs(data string) ([]string, bool) {
  if strings.Contains(data, HYPHEN) {
    slice := Split(data, HYPHEN)
    return []string{slice[0], slice[len(slice)-1]}, true
  } else if strings.Contains(data, COMMA) {
    return Extract(data), false
  }
  return []string{data}, false
}
=======
// formatting functions

package iof

import (
  "fmt"
  "log"
  "time"
  "bytes"
  "errors"
  "regexp"
  "slices"
  "reflect"
  "runtime"
  "strings"
  "strconv"
  "unicode"
  "crypto/md5"
  "encoding/hex"
  "encoding/json"
)

// get first row (dict) from resultset
func First(data []interface{}, structure []string) []string {
  return Row(0, data, structure)
}

// get a row (dict) from resultset["result"]
func Row(index int, data []interface{}, structure []string) []string {
  if index < len(data) {
    return Structify(data[index].(map[string]any), structure)
  } else {
    return nil
  }
}

// get all rows from resultset["result"]
func Array(data []interface{}, structure []string) [][]string {
  var rows [][]string
  for index := range data {
    row := Structify(data[index].(map[string]any), structure)
    rows = append(rows, row)
  }
  return rows
}

// get raw data (list of dicts) from resultset
func Result(data map[string]interface{}) []interface{} {
  return data[RESULT].([]interface{})
}

// get an item from structured slice
func Pick(data []string, id string, structure []string) string {
  if len(data) != len(structure) || !slices.Contains(structure, id) {
    log.Fatal(INVALIDSTRUCT())
  }
  for index := range structure {
    if structure[index] == id { return Nice(data[index]) }
  }
  return ""
}

// make data row object structured slice for output
func Structify(data map[string]any, structure []string) []string {
  clone := make([]string, len(structure))
  copy(clone, structure)
  for index, id := range structure {
    if item, exists := data[id]; exists {
      clone[index] = Nice(item)
    } else {
      log.Fatal(INVALIDSTRUCT())
    }
  }
  return clone
}

// get columns width of text matrix
func Widths(data [][]string) []int {
  cols := Wide(data)
  if cols == 0 { return []int{} }
  widths := []int{}
  for col := range cols {
    width := 0
    for row := range len(data) {
      w := len(data[row][col])
      if w > width { width = w }
    }
    widths = append(widths, width)
  }
  return widths
}

// count columns of 2d matrix
func Wide(data [][]string) int {
  if len(data) == 0 { return 0 }
  return len(data[0])
}

// generate "name" jsondata
func Jname(name string) string {
  return Stringify(map[string]interface{}{"name": name})
}

// implement javascript json.stringify
func Stringify(data map[string]interface{}) string {
  jdata, err := json.Marshal(data)
  if err != nil { log.Fatal(err) }
  return string(jdata)
}

// implement javascript json.parse
func Parse(sdata string) map[string]interface{} {
  var data map[string]interface{}
  err := json.Unmarshal([]byte(sdata), &data)
  if err != nil { log.Fatal(err) }
  return data
}

// assertion map of (string) to (any)
func Any(src map[string]string) map[string]any {
  if src == nil { return nil }
  dst := make(map[string]any, len(src))
  for key, val := range src { dst[key] = val }
  return dst
}

// assertion map of (any) to (string)
func FromAny(src map[string]any) map[string]string {
  if src == nil { return nil }
  dst := make(map[string]string, len(src))
  for key, val := range src { dst[key] = String(val) }
  return dst
}

// convert data to string
func String(data any) string {
  return fmt.Sprintf("%v", data)
}

// make nil as blank str
func Nice(data any) string {
  v := String(data)
  if v == "<nil>" { return "" }
  return  v
}

// convert string to integer
func Integer(data string) int {
  num, err := strconv.Atoi(data)
  if err == nil { return num }
  return 0
}

// convert string to float
func Float64(data string) float64 {
  num, err := strconv.ParseFloat(data, 64)
  if err == nil { return num }
  return 0
}

// use with Input() for none
func None(data any) (any, error) {
  return data, nil
}

// use with Input()
// convert data to string
func Str(data any) (any, error) {
  return any(String(data)), nil
}

// use with Input()
// convert data to integer
func Int(data any) (any, error) {
  num, err := strconv.Atoi(data.(string))
  if err == nil { return num, nil }
  return 0, err
}

// use with Input()
// convert data to float
func Float(data any) (any, error) {
  num, err := strconv.ParseFloat(data.(string), 64)
  if err == nil { return num, nil }
  return 0, err
}

// use with Input()
// convert data to Date()
func Cdate(data any) (any, error) {
  s, err := Date(String(data))
  return any(s), err
}

// use with Input()
// convert data to Time()
func Ctime(data any) (any, error) {
  s, err := Time(String(data))
  return any(s), err
}

// use with Input()
// convert data to bool
func Bool(data any) (any, error) {
  switch Lower(String(data)) {
  case "y", "yes", "true", "1": return true, nil
  default: return false, nil
  }
}

// convert string to bool
// to trigger some action
func Yes(data string) bool {
  v, _ := Bool(any(data))
  return v.(bool)
}

func StrCash(num string) string { return Cash(any(num)) }

// format float number string
// with thousands separator [,]
func Cash(num any) string {
  dec := DOT
  s := strings.TrimSpace(String(num))
  d := strings.Split(s, dec)
  if len(d) > 2 { return s }
  d1, err := strconv.Atoi(d[0])
  if True(err) { return s }
  number := money(d1)
  if len(d) == 1 { return number }
  d2, err := strconv.Atoi(d[1])
  if True(err) { return s }
  return number + dec + String(d2)
}

// format integer number string
// with thousands separator [,]
func money(n int) string {
  neg := "-"
  sign := ""
  in := strconv.Itoa(n)
  if string(in[0]) == neg {
    sign = neg
    in = in[1:]
  }
  var buf bytes.Buffer
  k := 0
  for i := len(in) - 1; i >= 0; i-- {
    buf.WriteByte(in[i])
    k++
    if k%3 == 0 && i != 0 {
      buf.WriteByte(',')
    }
  }
  return sign + reverse(buf.String())
}

func reverse(s string) string {
  runes := []rune(s)
  for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
    runes[i], runes[j] = runes[j], runes[i]
  }
  return string(runes)
}

// get md5 hash of string
func MD5(text string) string {
  data := []byte(text)
  hashBytes := md5.Sum(data)
  return hex.EncodeToString(hashBytes[:])
}

// get calling function name
// set level=0 for owner func
func Callup(level int) string {
  pc, _, _, ok := runtime.Caller(level + 1)
  if !ok { return "" }
  f := runtime.FuncForPC(pc)
  if f == nil { return "" }
  fn := strings.Split(f.Name(), DOT)
  return fn[len(fn)-1]
}

// capitalize first letter of string
func Capitalize(str string) string {
  var output []rune
  isWord := true
  for _, val := range str {
    if isWord && unicode.IsLetter(val) {
      output = append(output, unicode.ToUpper(val))
      isWord = false
    } else if !unicode.IsLetter(val) && val != '_' {
      isWord = true
      output = append(output, val)
    } else {
      output = append(output, val)
    }
  }
  return string(output)
}

func Lower(str string) string {
  return strings.ToLower(str)
}

func Upper(str string) string {
  return strings.ToUpper(str)
}

// parse menu option to function id
func Optionize(option string) string {
  // need some stuffs here
  return Lower(option)
}

// parse menu option to agent func
func Agentize(option string) string {
  return Capitalize(Optionize(option))
}

// get [map] object attr as string
func Get(data map[string]string, attr string) string {
  if item, ok := data[attr]; ok { return item }
  return ""
}

// get [map] object attr as any type
func Pull(data map[string]any, attr string) any {
  if item, ok := data[attr]; ok { return item }
  return nil
}

func IsStr(value any) bool {
  _, ok := value.(string)
  return ok
}

func IsBool(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Bool
}

func IsInt(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Int
}

func IsFloat(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Float64
}

func IsMap(data any) bool {
  v := reflect.ValueOf(data)
  return v.Kind() == reflect.Map
}

func IsArray(data any) bool {
  v := reflect.ValueOf(data)
  return (v.Kind() == reflect.Array || v.Kind() == reflect.Slice)
}

// check two vars equal
func Matched(var1 any, var2 any) bool {
  return True(var2) && var2 == var1
}

// check var as boolean
func False(data any) bool {
  return !True(data)
}

// check var as boolean
func True(data any) bool {
  if data == nil { return false }
  if IsBool(data) { return data.(bool) }
  if IsStr(data) && data == "" { return false }
  if (IsInt(data) || IsFloat(data)) && data == 0 { return false }
  if (IsMap(data) || IsArray(data)) && reflect.ValueOf(data).Len() == 0 { return false }
  return true
}

// retrieve maps in data slice where specific key matches given term
func Search(data []map[string]string, key, term string) []map[string]string {
  var results []map[string]string
  for _, m := range data {
    if val, ok := m[key]; ok && val == term {
      results = append(results, m)
    }
  }
  return results
}

// retrieve single first map of Search() function
func Find(data []map[string]string, key, term string) map[string]string {
  results := Search(data, key, term)
  if len(results) > 0 { return results[0] }
  return map[string]string{}
}

// get valid menu option index+1
func Optioned(choice string, total int) int {
  num := Integer(choice)
  if num < 1 || num > total { return 0 }
  return num
}

// get valid menu multi-options index+1
func MultiOptioned(choices []string, total int) []int {
  var options []int
  for _, choice := range choices {
    num := Optioned(choice, total)
    if num == 0 { return []int{} }
    options = append(options, num)
  }
  return options
}

// validate agent function
func Valid(funcname string) bool {
  return reflect.ValueOf(Standalone{}).MethodByName(funcname).IsValid()
}

var Agented = Valid // alias

// apply bool to db bool
func DbBool(data bool) int {
  if data { return 1 }
  return 0
}

// Convert str bool to yes/no
func StrBool(data string) string {
  if Include(any(data), []any{"<nil>", nil, "false", false, "0", 0, ""}) { return LABELS["NO"] }
  return LABELS["YES"]
}

// shorten sensitive string
func StrZip(text string) string {
  if len(text) < 5 { return THREEDOT }
  return text[0:5] + THREEDOT
}

// convert db time to date
func StrDate(time string) string {
  if False(time) || Snil(time) || len(time) < 10 { return "" }
  return time[0:10]
}

// Convert uk date to rfc3339
func Date(data string) (string, error) {
  const layout = "2/1/2006"
  t, err := time.Parse(layout, data)
  if True(err) { return "", errors.New(INVALIDINPUT()) }
  return t.Format(time.RFC3339)[0:10], nil
}

// Adjust hour:minute as rfc3339
func Time(data string) (string, error) {
  const layout = "2006-01-02 15:4"
  data = "2006-01-02 " + data
  t, err := time.Parse(layout, data)
  if True(err) { return "", errors.New(INVALIDINPUT()) }
  return t.Format(time.RFC3339)[11:16], nil
}

// split string to array
func Split(data string, separator string) []string {
  slice := strings.Split(data, separator)
  for i, val := range slice { slice[i] = strings.TrimSpace(val) }
  return slice
}

// extract string to array
func Extract(data string) []string {
  slice := Split(data, COMMA)
  for i, val := range slice { slice[i] = strings.TrimSpace(val) }
  return slice
}

// extract string to int interval
func Interval(data string) []int {
  slice := Split(data, HYPHEN)
  return []int{Integer(slice[0]), Integer(slice[len(slice)-1])}
}

// extract string to array of int
func Integers(data string) []int {
  slice := Extract(data)
  inums := make([]int, len(slice))
  for i, val := range slice { inums[i] = Integer(val) }
  return inums
}

// clone python f-string with [%v] values
func Fstr(rawstr string, agrs ...any) string {
  return fmt.Sprintf(rawstr, agrs...)
}

// clear all puncs in string
func Escape(data string) string {
  reg := regexp.MustCompile("[^\\p{L}\\p{N}]+")
  return reg.ReplaceAllString(data, "")
}

// reformat input string for names and ids
func Form(data string, format string) string {
  forms := Extract(format)
  for _, form := range forms {
    switch Lower(form) {
    case "upper", "up", "u":
      data = Upper(data)
    case "lower", "low", "l":
      data = Lower(data)
    case "clear", "clr", "c":
      data = Escape(data)
    }
  }
  return data
}

// check integer in range
func Ranged(val int, data []int) bool {
  return len(data) == 2 && val >= data[0] && val <= data[1]
}

// check integer in slice
func In(val int, data []int) bool {
  return slices.Contains(data, val)
}

// check any value in slice
func Include(val any, data []any) bool {
  return slices.Contains(data, val)
}

// convert server response interface to map
func Map(data []interface{}) []map[string]any {
  result := make([]map[string]any, len(data))
  for index, val := range data {
    if mapval, ok := val.(map[string]any); ok {
      result[index] = mapval
    } else {
      return nil
    }
  }
  return result
}

func Now() string { return time.Now().Format(time.RFC3339) }

func Update(data map[string]interface{}, key string, value any) { data[key] = value }
func Default(data map[string]interface{}, key string, value any) { if False(Pull(data, key)) { data[key] = value } }

func Type(data any) string { return fmt.Sprintf("%T", data) } // get var type as string
func Snil(data string) bool { return data == "<nil>" } // caution: check string as nil
func Nil(data any) bool { return data == nil } // check var as nil

// extract str to int range
func Ints(data string) ([]int, bool) {
  if strings.Contains(data, HYPHEN) {
    return Interval(data), true
  } else if strings.Contains(data, COMMA) {
    return Integers(data), false
  }
  return []int{Integer(data)}, false
}

// extract str to str range
func Strs(data string) ([]string, bool) {
  if strings.Contains(data, HYPHEN) {
    slice := Split(data, HYPHEN)
    return []string{slice[0], slice[len(slice)-1]}, true
  } else if strings.Contains(data, COMMA) {
    return Extract(data), false
  }
  return []string{data}, false
}
