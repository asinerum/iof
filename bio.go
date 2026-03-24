// basic io lib

package iof

import (
  "os"
  "fmt"
  "log"
  "bufio"
  "slices"
  "reflect"
  "runtime"
  "strconv"
  "strings"
  "syscall"
  "os/exec"
  "encoding/json"
  "text/tabwriter"

  "golang.org/x/term"
  "github.com/fatih/color"
)

// process user menu option with or without args
func Action(menuid string, args map[string]interface{}) {
  // loop to activate menu eternally
  outer:
  for {
    option := Option(menuid, false).(string)
    if option == QUIT { break }
    agentfunc := Agentize(option)
    if Agented(agentfunc) {
      Agent(agentfunc, args)
    } else {
      hidtask := Find(HIDTASKS, "task", option)
      var inputs map[string]interface{}
      for {
        inputs = GetInput(option)
        if True(hidtask) && !Matched(Pull(inputs, hidtask["var1"]), Pull(inputs, hidtask["var2"])) {
          fmt.Println()
          Error(hidtask["fail"])
          continue
        }
        fmt.Println()
        DataTitle(TEXTS["MSG_input_data"])
        fmt.Println()
        if True(hidtask) {
          Out(map[string]any{"input": hidtask["note"]})
        } else {
          Out(inputs)
        }
        fmt.Println()
        answer := Prompt()
        if answer == YES { break }
        if answer == CANCEL { continue outer }
      }
      result := DirectCall(option, Stringify(inputs))
      if Failed(result) {
        fmt.Println()
        Error(Err(result))
      } else {
        if Empty(result) {
          fmt.Println()
          Warning(NODATAFOUND())
        } else {
          dat := Map(Result(result))
          if False(dat) {
            Error(INVALIDDATA())
            return
          }
          DataReform(option, dat)
          retmod := Get(RETURNS[option], "return")
          if retmod == "item" {
            Echo(option, dat[0], "Result Item")
          } else if retmod == "rows" {
            TabEcho(option, dat, "Result Set")
          } else if True(retmod) {
            fmt.Println()
            Warning(LABELS[retmod])
          } else {
            fmt.Println()
            Out(result)
          }
        }
      }
    }
  }
}

// api post
func Post() {
  fmt.Println()
  Warning(Capitalize(TEXTS["MSG_api_direct_post"]))
  var path string
  for {
    fmt.Println()
    path = GetStr(true, TEXTS["MSG_api_json_file"])
    fmt.Println()
    if Confirm() { break }
  }
  data := Json(path)
  function := String(Pull(data, "function"))
  jsondata := Stringify(Pull(data, "jsondata").(map[string]any))
  fmt.Println()
  Out(DirectCall(function, jsondata))
}

// api call
func Direct() {
  fmt.Println()
  Warning(Capitalize(TEXTS["MSG_api_direct_call"]))
  var function string
  var jsondata string
  for {
    fmt.Println()
    function = GetStr(true, TEXTS["MSG_api_function"])
    fmt.Println()
    jsondata = GetStr(true, TEXTS["MSG_api_jsondata"])
    fmt.Println()
    if Confirm() { break }
  }
  fmt.Println()
  Out(DirectCall(function, jsondata))
}

// call a function by name and parse map arguments
func Agent(gofunc string, args map[string]interface{}) {
  vt := reflect.ValueOf(Standalone{})
  vm := vt.MethodByName(gofunc)
  if args == nil {
    vm.Call([]reflect.Value{})
  } else {
    vm.Call([]reflect.Value{reflect.ValueOf(args)})
  }
}

// run a menu then get users selection
func Option(menuid string, multi bool) any {
  menu := Menu()[menuid]
  return Select(multi, LABELS[menuid], menu, LABELS)
}

// create menu from map of functions id and name
func Select(multi bool, title string, options []string, labels map[string]string) any {
  reader := bufio.NewReader(os.Stdin)
  fmt.Println()
  MenuTitle(title)
  fmt.Println()
  length := len(options) + 1
  for index, option := range options {
    MenuLine(String(index + 1) + DOT + SPACE + labels[option])
  }
  MenuLine(String(length) + DOT + SPACE + MSG_exit())
  fmt.Println()
  var selected string
  var selects []string
  for {
    fmt.Print(MSG_choice(), COLON, SPACE)
    input, _ := reader.ReadString(CR)
    choice := strings.TrimSpace(input)
    if !multi {
      num := Optioned(choice, length)
      if num == 0 { continue }
      if num == length {
        selected = QUIT
      } else {
        selected = options[num - 1]
      }
      return selected
    } else {
      nums := MultiOptioned(Extract(choice), length)
      if False(nums) { continue }
      for _, num := range nums {
        selects = append(selects, options[num - 1])
      }
      return selects
    }
  }
}

// print colored text via agent
func Paint(target string, text string) {
  Agent(OUTPUTS["colors"][target], map[string]interface{}{"text": text})
}

func (c Standalone) Red(args map[string]interface{}) { color.Red(args["text"].(string)) }
func (c Standalone) Green(args map[string]interface{}) { color.Green(args["text"].(string)) }
func (c Standalone) Blue(args map[string]interface{}) { color.Blue(args["text"].(string)) }
func (c Standalone) Yellow(args map[string]interface{}) { color.Yellow(args["text"].(string)) }
func (c Standalone) Cyan(args map[string]interface{}) { color.Cyan(args["text"].(string)) }
func (c Standalone) Magenta(args map[string]interface{}) { color.Magenta(args["text"].(string)) }
func (c Standalone) White(args map[string]interface{}) { color.White(args["text"].(string)) }
func (c Standalone) Black(args map[string]interface{}) { color.Black(args["text"].(string)) }

// output table from data set
func Table(head []string, data [][]string, pad int) {
  wide := Wide(data)
  if len(head) != wide { log.Fatal(INVALIDINPUT()) }
  widths := Widths(append(data, head))
  bar := slices.Clone(head)
  for i := range wide { bar[i] = strings.Repeat(HYPHEN, widths[i]+pad) }
  w := Writer(2)
  fmt.Fprintln(w, Tab(head))
  fmt.Fprintln(w, Tab(bar))
  for i := range data { fmt.Fprintln(w, Tab(data[i])) }
  w.Flush()
}

// output table from database resultset
func Tablet(data []map[string]any, cols []string, heads map[string]string) {
  if False(data) {
    fmt.Println(NODATAFOUND())
    return
  }
  h := Header(heads, cols)
  var d [][]string
  for index := range data {
    var r []string
    for _, col := range cols {
      val, ok := data[index][col]
      if !ok { log.Fatal(INVALIDSTRUCT()) }
      r = append(r, String(val))
   }
   d = append(d, r)
  }
  Table(h, d, 2)  
}

// output structured column from data row
func Col(data map[string]any, structure []string, heads []string) {
  if heads == nil { heads = structure }
  w :=  Writer(2)
  dat := Structify(data, structure)
  for index := range structure {
    fmt.Fprintln(w, Tab([]string{heads[index], dat[index]}))
  }
  w.Flush()
}

// output raw json data as a column with specific structure and headers
func Column(id string, data map[string]any, structing []string, heading []string, defheader string) {
  fmt.Println()
  DataTitle(Captioning(id, HEADER, defheader))
  fmt.Println()
  Col(data, structing, heading)
}

// output raw json data as a column using given structure
func Echo(id string, data map[string]any, defheader string) {
  Column(id, data, Structing(id), Heading(id), defheader)
}

// output raw json data as a table using given structure
func TabEcho(id string, data []map[string]any, defheader string) {
  fmt.Println()
  DataTitle(Captioning(id, HEADER, defheader))
  fmt.Println()
  Tablet(data, Structing(id), Heads(id))
}

// output preserved data list
func Dat(data map[string]string, structure []string) {
  w :=  Writer(2)
  for _, id := range structure { fmt.Fprintln(w, Tab([]string{id, data[id]})) }
  w.Flush()
}

// output given data list
func List(id string, defheader string) {
  datalist, structing, _ := DataList(id)
  fmt.Println()
  MenuLine(Capitalize(LABELS["data_" + id]))
  fmt.Println()
  Dat(datalist, structing)
}

// standard tabwriter.NewWriter
func Writer(pad int) *tabwriter.Writer { // this [pad] should be 2
  return tabwriter.NewWriter(os.Stdout, 0, 0, pad, ' ', tabwriter.AlignRight)
}

// format table row from slice
func Tab(data []string) string {
  for i, v := range data { data[i] = Nice(v) }
  return strings.Join(data[:], TAB) + TAB
}

// output raw json data
func Out(data map[string]any) {
  jdata, err := json.MarshalIndent(data, "", "  ")
  if err != nil { log.Fatal(err) }
  fmt.Println(string(jdata))
}

// clear screen
func Cls() {
  var cmd *exec.Cmd
  if runtime.GOOS == "windows" {
    cmd = exec.Command("cmd", "/c", "cls")
  } else {
    cmd = exec.Command("clear")
  }
  cmd.Stdout = os.Stdout
  cmd.Run()
}

// input password from stdin with mask
func Password(prompt string, errinput string) string {
  fmt.Println(prompt, BACKCOLON)
  for {
    bytes, err := term.ReadPassword(int(syscall.Stdin))
    if True(err) {
      fmt.Println(errinput)
      continue
    }
    pwd := string(bytes)
    return strings.TrimSpace(pwd)
  }
}

// raw input from stdin
func Scan(prompt string, hprompt string, errinput string) string {
  fmt.Println(prompt, BACKCOLON)
  if True(hprompt) { fmt.Println("(" + hprompt + ")") }
  scanner := bufio.NewScanner(os.Stdin)
  for !scanner.Scan() {
    fmt.Println(errinput)
    continue
  }
  return strings.TrimSpace(scanner.Text())
}

// input from stdin with value checking
func Input(require bool, check func(any) (any, error), prompt string, hprompt string, defval any, errinput string, errvalue string) any {
  prompt = strings.ReplaceAll(prompt, "(?)", "(" + String(defval) + ")")
  for {
    input := Scan(prompt, hprompt, errinput)
    if input == "" {
      if require { continue }
      return nil // caution
    }
    if val, err := check(any(input)); err == nil { return val }
    if True(errvalue) { fmt.Println(errvalue) } // warn when needed
  }
}

// user enters an integer with api check
func GetInt(require bool, prompt string, hprompt string, defval int, errvalue string, checkfunc func(any) (any, error)) int {
  input := Input(require, checkfunc, prompt, hprompt, any(defval), ERRORINPUT(), errvalue) // can be nil
  val, err := strconv.Atoi(String(input))
  if True(err) { return defval }
  return val
}

// user enters a float with api check
func GetFloat(require bool, prompt string, hprompt string, defval float64, errvalue string) float64 {
  input := Input(require, None, prompt, hprompt, any(defval), ERRORINPUT(), errvalue) // can be nil
  val, err := strconv.ParseFloat(String(input), 64)
  if True(err) { return defval }
  return val
}

func Confirm() bool { return GetBool(TEXTS["MSG_confirm"], false) }

func Prompt() string { return GetOne(false, TEXTS["MSG_confirm_prompt"], Extract(PROMPT), NO) }

// user enters a boolean value
func GetBool(prompt string, defval bool) bool {
  val := Input(false, Bool, prompt, "", any(defval), ERRORINPUT(), "")
  if Nil(val) { return defval }
  return val.(bool)
}

// user enters a date
func GetDate(require bool, prompt string, defval string) string {
  val := Input(require, Cdate, prompt, "", any(defval), ERRORINPUT(), "")
  if False(val) { return defval }
  return String(val)
}

// user enters a time [hour:min]
func GetTime(require bool, prompt string, defval string) string {
  val := Input(require, Ctime, prompt, "", any(defval), ERRORINPUT(), "")
  if False(val) { return defval }
  return String(val)
}

// user enters a string with api check
func GetString(require bool, prompt string, hprompt string, defval string, errvalue string, checkfunc func(any) (any, error)) string {
  val := Input(require, checkfunc, prompt, hprompt, any(defval), ERRORINPUT(), errvalue)
  if False(val) { return defval }
  return String(val)
}

// user enters a simple string
func GetStr(require bool, prompt string) string {
  val := Input(require, Str, prompt, "", nil, ERRORINPUT(), INVALIDINPUT())
  if False(val) { return "" }
  return String(val)
}

// user enters a simple choice from options
func GetOne(require bool, prompt string, values []string, defval string) string {
  val := Lower(String(Input(require, None, prompt, "", defval, ERRORINPUT(), INVALIDINPUT())))
  if slices.Contains(values, val) { return val }
  return defval
}

// user enters a password
func GetPwd(prompt string) string {
  return Password(prompt, ERRORINPUT())
}

// user enters credentials
func GetLogin(endpoint string) {
  username := GetStr(true, MSG_login_name())
  password := MD5(GetPwd(MSG_login_password()))
  err := Login(endpoint, username, password)
  if err != nil { log.Fatal(err) }
  SERVERENDPOINT = endpoint
}

func DataTitle(text string) { Paint("DataTitle", text) }
func MenuTitle(text string) { Paint("MenuTitle", text) }
func MenuLine(text string) { Paint("MenuLine", text) }
func Warning(text string) { Paint("Warning", text) }
func Error(text string) { Paint("Error", text) }

// load yaml data instead of server call
func DataList(id string) (map[string]string, []string, []string) {
  datalist := map[string]string{}
  structing := []string{}
  heading := []string{}
  for _, data := range DATA[id] {
    id := data["id"]
    name := data["name"]
    datalist[id] = name
    structing = append(structing, id)
    heading = append(heading, name)
  }
  return datalist, structing, heading
}

// reformat time and number strings as required
func DataReform(apifunc string, resultset []map[string]any) {
  for index, row := range resultset {
    for key, val := range row {
      if reform, ok := REFORMS[apifunc][key]; ok {
        if refunc, ok := StrFunctions[reform]; ok {
          resultset[index][key] = refunc(String(val))
        }
      }
    }
  }
}

// recalculate database result offset for server calling
func Offset(inputs map[string]interface{}, deflimit int, defoffset int, page string, limit string, offset string) {
  if False(page) { page = PAGE }
  if False(limit) { limit = LIMIT }
  if False(offset) { offset = OFFSET }
  if False(deflimit) { deflimit = DEFLIMIT }
  if False(defoffset) { defoffset = DEFOFFSET }
  Default(inputs, limit, deflimit)
  Default(inputs, offset, defoffset)
  Update(inputs, limit, max(inputs[limit].(int), 0))
  Update(inputs, offset, max(inputs[offset].(int), 0))
  no := Pull(inputs, page)
  var pageno int
  if True(no) {
    pageno = Integer(String(no))
  } else {
    pageno = inputs[offset].(int)
  }
  Update(inputs, offset, max(inputs[limit].(int)*(pageno-1), 0))
}

// recalculate offset value for database resulset paginating
func Paginate(inputs map[string]interface{}) { Offset(inputs, DEFLIMIT, DEFOFFSET, "", "", "") }

// this bool indicates that database is to be updated after api calling
func ToUpdate(apifunc string) bool { return Yes(Get(RETURNS[apifunc], "update")) }

// this bool indicates that server calling failed
func Failed(result map[string]interface{}) bool { return True(result[ERROR]) }

// this bool indicates that server database resulset contains nothing
func Empty(result map[string]interface{}) bool { return False(Result(result)) }

// get server calling error message
func Err(result map[string]interface{}) string { return String(result[ERROR]) }

// check api func for own client/zone
func SelfZone(apifunc string) bool { return apifunc[0:3] == "us_" }

// check api func for own user
func SelfUser(apifunc string) bool { return apifunc[0:3] == "me_" }

// check api func if updates zone-id
func UpZone(apifunc string) bool { return apifunc[0:3] == "zi_" }

// check api func if updates user-id
func UpUser(apifunc string) bool { return apifunc[0:3] == "ui_" }

// make data input for calling api service
// need further process for int[] before posting
func GetInput(apifunc string) map[string]interface{} {
  structure := Instruct()[apifunc]
  configs := Inscript()[apifunc]
  update := ToUpdate(apifunc)
  fmt.Println()
  DataTitle(Capitalize(LABELS[apifunc]))
  input := map[string]interface{}{}
  for _, val := range structure {
    prompt := Get(configs[val], "prompt")
    hprompt := Get(configs[val], "hprompt")
    values := Get(configs[val], "values")
    defval := Get(configs[val], "defval")
    format := Get(configs[val], "format")
    forbid := Get(configs[val], "forbid")
    require := Yes(Get(configs[val], "require"))
    datalist := Yes(Get(configs[val], "datalist"))
    var checkfunc func(string, string) (func(any) (any, error))
    var errvalue string
    var checkapi string
    noexistapi := Get(configs[val], "noexistapi")
    noexistmsg := Get(configs[val], "noexistmsg")
    existapi := Get(configs[val], "existapi")
    existmsg := Get(configs[val], "existmsg")
    if True(noexistapi) {
      checkfunc = CheckNotExist
      checkapi = noexistapi
      errvalue = noexistmsg
    } else if True(existapi) {
      checkfunc = CheckExist
      checkapi = existapi
      errvalue = existmsg
    } else {
      checkfunc = CheckNothing
      checkapi = ""
      errvalue = ""
    }
    if datalist { List(apifunc, "[given data]") }
    fmt.Println()
    for {
      switch Get(configs[val], "type") {
      case "int":
        input[val] = GetInt(require, prompt, hprompt, Integer(defval), errvalue, checkfunc(checkapi, val))
        if True(values) && !Ranged(input[val].(int), Interval(values)) && !In(input[val].(int), Integers(values)) { continue }
        if True(forbid) && In(input[val].(int), Integers(forbid)) { continue }
      case "float":
        input[val] = GetFloat(require, prompt, hprompt, Float64(defval), errvalue)
      case "bool":
        input[val] = DbBool(GetBool(prompt, Yes(defval)))
      case "date":
        input[val] = GetDate(require, prompt, defval)
      case "time":
        input[val] = GetTime(require, prompt, defval)
      case "pwd":
        input[val] = GetPwd(prompt)
      case "str":
        input[val] = Form(GetString(require, prompt, hprompt, defval, errvalue, checkfunc(checkapi, val)), format)
        if True(forbid) && Include(any(input[val]), []any{Extract(forbid)}) { continue }
      case "str[]":
        input[val] = Extract(GetString(require, prompt, hprompt, defval, errvalue, CheckNone()))
      case "int[]":
        input[val] = Integers(GetString(require, prompt, hprompt, defval, errvalue, CheckNone()))
      default:
        return map[string]interface{}{}
      }
      break
    }
  }
  if SelfZone(apifunc) { input[IDATTR] = Integer(Getzone()) }
  if SelfUser(apifunc) { input[NAMEATTR] = Getuser() }
  if UpZone(apifunc) { Zone(input) }
  if UpUser(apifunc) { User(input) }
  if !update { Paginate(input) }
  input["worktime"] = Now()
  return input
}
