// data io lib

package iof

import (
  "os"
  "fmt"
  "log"
  "time"
  "errors"
  "crypto/tls"
  "io/ioutil"
  "net/url"
  "net/http"
  "net/http/cookiejar"
  "encoding/json"
)

// just check nothing
func CheckNone(args ...any) func(any) (any, error) {
  return func(value any) (any, error) { return value, nil }
}

func CheckNothing(checkapi string, colid string) func(any) (any, error) {
  return CheckNone()
}

func CheckNotExist(checkapi string, colid string) func(any) (any, error) {
  return Exist(checkapi, colid, false)
}

func CheckExist(checkapi string, colid string) func(any) (any, error) {
  return Exist(checkapi, colid, true)
}

// check if resultset is not empty from api call
func Exist(checkapi string, colid string, trueifexist bool) func(any) (any, error) {
  inner := func(value any) (any, error) {
    result := Call(SERVERENDPOINT, "", checkapi, Stringify(map[string]interface{}{colid: value}))
    if False(result[ERROR]) && True(Result(result)) {
      if trueifexist { return value, nil }
      return nil, errors.New(INVALIDINPUT())
    } else {
      if trueifexist { return nil, errors.New(INVALIDINPUT()) }
      return value, nil
    }
  }
  return inner
}

// use shortened Call() to call for server data as an agent function
func DirectCall(functionid string, jsondata string) map[string]interface{} {
  return Call(SERVERENDPOINT, "", functionid, jsondata)
}

// call for data function and get response from server
func Call(endpoint string, moduleid string, functionid string, jsondata string) map[string]interface{} {
  input := map[string]interface{}{
    "module": moduleid,
    "function": functionid,
    "jsondata": jsondata,
  }
  data, err := Server(endpoint + "/call", Stringify(input))
  if err != nil { log.Fatal(err) }
  return data
}

// send request to server to get raw data
func (c Standalone) Ask(args map[string]interface{}) {
  route := String(args["route"])
  sparam := String(args["sparam"])
  ApiOut(Endpoint(route), sparam)
}

// global instance of Standalone.Ask
func Ask(route string, sparam string) {
  Standalone{}.Ask(map[string]interface{}{"route": route, "sparam": sparam})
}

// run cli command with one argument
// command "hello", "init", "net" etc
func Do(command string, param string) {
  // get route/command id from task.yaml
  task := Find(TOPTASKS, "cmd", command)
  if True(task) {
    if param == DOT {
      Agent(task["agent"], nil)
    } else {
      var sparam string
      arg := Get(task, "arg") // route/command arg from task.yaml
      if True(arg) { sparam = Stringify(map[string]any{arg: param}) }
      Agent(task["agent"], map[string]interface{}{"route": task["route"], "sparam": sparam})
    }
  } else {
    fmt.Println(NODATAFOUND())
  }
}

// init client/zone
// route: init/zone
// route: init/accounting
// parameter: zoneid [int]
func Zoner(route string) {
  fmt.Println()
  zoneid := GetInt(true, TEXTS["MSG_client_id"], "", 0, TEXTS["MSG_client_id_not_exist"], CheckExist("select_zone_by_id_or_name", "id"))
  // no printing new line here: fmt.Println()
  Ask(route, Fstr(`{"zoneid":%v}`, zoneid))
}

// output raw data (got from server) as a column
func Print(id string, endpoint string, sparam string) {
  Echo(id, ApiCall(endpoint, sparam), "[server response]")
}

// output/print raw data from server
func ApiOut(endpoint string, sparam string) {
  fmt.Println()
  Out(ApiCall(endpoint, sparam))
}

// get raw data from server
func ApiCall(endpoint string, sparam string) map[string]any {
  data, err := Server(endpoint, sparam)
  if err != nil { log.Fatal(err) }
  return data
}

// get endpoint url from route id
func Endpoint(route string) string {
  return SERVERENDPOINT + "/" + route
}

// login to server and start working session
func Login(endpoint string, username string, password string) error {
  input := map[string]interface{}{
    "username": username,
    "password": password,
  }
  data, err := Server(endpoint + "/login", Stringify(input))
  if err != nil { log.Fatal(err) }
  ACCESSTOKEN = String(data[TOKEN])
  if ACCESSTOKEN == ERROR {
    Deluser()
    Delzone()
    Delzadm()
    return fmt.Errorf("%s", String(data[TOKENTYPE]))
  } else {
    Setuser(String(data[USERATTR]))
    Setzone(String(data[ZONEATTR]))
    Setzadm(String(data[ZADMATTR]))
    r := String(data[USERATTR])
    fmt.Println(AUTHORIZED(), BACKCOLON, r)
    if True(SaveToken(ACCESSTOKEN)) { Error(ERRORSAVE()) }
    return nil
  }
}

// set/delete/get user name of working session
func Setuser(name string) {os.Setenv(SESSIONUSER, name)}
func Deluser() {os.Setenv(SESSIONUSER, "")}
func Getuser() string {return os.Getenv(SESSIONUSER)}

// set/delete/get client id of working session
func Setzone(id string) {os.Setenv(SESSIONZONE, id)}
func Delzone() {os.Setenv(SESSIONZONE, "")}
func Getzone() string {return os.Getenv(SESSIONZONE)}

// set/delete/get user admin privilege of working session
func Setzadm(status string) {os.Setenv(SESSIONZADM, status)}
func Delzadm() {os.Setenv(SESSIONZADM, "")}
func Getzadm() string {return os.Getenv(SESSIONZADM)}

// call server with json-string parameter using both methods
func Server(endpoint string, sparam string) (map[string]interface{}, error) {
  if APIACCESS == OAUTH2 {
    return ServerDo(endpoint, sparam, ACCESSTOKEN)
  } else {
    return ServerGet(endpoint, sparam)
  }
}

// call server with json-string parameter using stored token
func ServerDo(endpoint string, sparam string, access_token string) (map[string]interface{}, error) {
  req, err := http.NewRequest("GET", Url(endpoint, sparam), nil)
  if True(err) { return nil, err }
  req.Header.Set("Authorization", "Bearer " + access_token)
  resp, err := app_session.Do(req)
  if True(err) { return nil, err }
  defer resp.Body.Close()
  return Response(resp)
}

// call server with json-string parameter using cookie
func ServerGet(endpoint string, sparam string) (map[string]interface{}, error) {
  resp, err := app_session.Get(Url(endpoint, sparam))
  if True(err) { return nil, err }
  defer resp.Body.Close()
  return Response(resp)
}

// process server response
func Response(resp *http.Response) (map[string]interface{}, error) {
  body, err := ioutil.ReadAll(resp.Body)
  if True(err) { return nil, err }
  var data map[string]interface{}
  err = json.Unmarshal(body, &data)
  if True(err) { return nil, err }
  return data, nil
}

// parse url string for get request
func Url(endpoint string, sparam string) string {
  if sparam == "" { return endpoint }
  params := url.Values{}
  for key, value := range Parse(sparam) { params.Set(key, String(value)) }
  return fmt.Sprintf("%s?%s", endpoint, params.Encode())
}

func Cookie() {
  if APIACCESS == OAUTH2 {
    // use saved access token
    app_session = &http.Client {
      Transport: &http.Transport {
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
      },
    }
  } else {
    // use cookie to store session
    jar, err := cookiejar.New(nil)
    if err != nil { log.Fatal(err) }
    app_session = &http.Client{
      Jar: jar,
      Timeout: 10 * time.Second,
    }
  }
}

func SaveToken(token string) error {
  data, err := json.MarshalIndent(map[string]string{TOKEN: token}, "", "")
  if True(err) { return err }
  return ioutil.WriteFile(TOKENJSON, data, 0644)
}

func LoadToken() (string, error) {
  data, err := ioutil.ReadFile(TOKENJSON)
  if True(err) { return "", err }
  var token map[string]string
  err = json.Unmarshal(data, &token)
  if True(err) { return "", err }
  return Get(token, TOKEN), nil
}

// get user id by name
func Uid(name string) int {
  data := Result(DirectCall("select_user_id_by_name", Jname(name)))
  if False(data) { return -1 }
  return Integer(FromAny(data[0].(map[string]any))[IDATTR])
}

// get client id by name
func Zid(name string) int {
  data := Result(DirectCall("select_zone_id_by_name", Jname(name)))
  if False(data) { return -1 }
  return Integer(FromAny(data[0].(map[string]any))[IDATTR])
}

// update input["user"] from ["name"]
func User(input map[string]interface{}) {
  if Pull(input, USERATTR) == nil {
    input[USERATTR] = Uid(String(Pull(input, NAMEATTR)))
  }
}

// caution to ambiguous ["name"]
func Zone(input map[string]interface{}) {
  if Pull(input, ZONEATTR) == nil {
    input[ZONEATTR] = Zid(String(Pull(input, NAMEATTR)))
  }
}
=======
// data io lib

package iof

import (
  "os"
  "fmt"
  "log"
  "time"
  "errors"
  "crypto/tls"
  "io/ioutil"
  "net/url"
  "net/http"
  "net/http/cookiejar"
  "encoding/json"
)

// just check nothing
func CheckNone(args ...any) func(any) (any, error) {
  return func(value any) (any, error) { return value, nil }
}

func CheckNothing(checkapi string, colid string) func(any) (any, error) {
  return CheckNone()
}

func CheckNotExist(checkapi string, colid string) func(any) (any, error) {
  return Exist(checkapi, colid, false)
}

func CheckExist(checkapi string, colid string) func(any) (any, error) {
  return Exist(checkapi, colid, true)
}

// check if resultset is not empty from api call
func Exist(checkapi string, colid string, trueifexist bool) func(any) (any, error) {
  inner := func(value any) (any, error) {
    result := Call(SERVERENDPOINT, "", checkapi, Stringify(map[string]interface{}{colid: value}))
    if False(result[ERROR]) && True(Result(result)) {
      if trueifexist { return value, nil }
      return nil, errors.New(INVALIDINPUT())
    } else {
      if trueifexist { return nil, errors.New(INVALIDINPUT()) }
      return value, nil
    }
  }
  return inner
}

// use shortened Call() to call for server data as an agent function
func DirectCall(functionid string, jsondata string) map[string]interface{} {
  return Call(SERVERENDPOINT, "", functionid, jsondata)
}

// call for data function and get response from server
func Call(endpoint string, moduleid string, functionid string, jsondata string) map[string]interface{} {
  input := map[string]interface{}{
    "module": moduleid,
    "function": functionid,
    "jsondata": jsondata,
  }
  data, err := Server(endpoint + "/call", Stringify(input))
  if err != nil { log.Fatal(err) }
  return data
}

// send request to server to get raw data
func (c Standalone) Ask(args map[string]interface{}) {
  route := String(args["route"])
  sparam := String(args["sparam"])
  ApiOut(Endpoint(route), sparam)
}

// global instance of Standalone.Ask
func Ask(route string, sparam string) {
  Standalone{}.Ask(map[string]interface{}{"route": route, "sparam": sparam})
}

// run cli command with one argument
// command "hello", "init", "net" etc
func Do(command string, param string) {
  // get route/command id from task.yaml
  task := Find(TOPTASKS, "cmd", command)
  if True(task) {
    if param == DOT {
      Agent(task["agent"], nil)
    } else {
      var sparam string
      arg := Get(task, "arg") // route/command arg from task.yaml
      if True(arg) { sparam = Stringify(map[string]any{arg: param}) }
      Agent(task["agent"], map[string]interface{}{"route": task["route"], "sparam": sparam})
    }
  } else {
    fmt.Println(NODATAFOUND())
  }
}

// init client/zone
// route: init/zone
// route: init/accounting
// parameter: zoneid [int]
func Zoner(route string) {
  fmt.Println()
  zoneid := GetInt(true, TEXTS["MSG_client_id"], "", 0, TEXTS["MSG_client_id_not_exist"], CheckExist("select_zone_by_id_or_name", "id"))
  // no printing new line here: fmt.Println()
  Ask(route, Fstr(`{"zoneid":%v}`, zoneid))
}

// output raw data (got from server) as a column
func Print(id string, endpoint string, sparam string) {
  Echo(id, ApiCall(endpoint, sparam), "[server response]")
}

// output/print raw data from server
func ApiOut(endpoint string, sparam string) {
  fmt.Println()
  Out(ApiCall(endpoint, sparam))
}

// get raw data from server
func ApiCall(endpoint string, sparam string) map[string]any {
  data, err := Server(endpoint, sparam)
  if err != nil { log.Fatal(err) }
  return data
}

// get endpoint url from route id
func Endpoint(route string) string {
  return SERVERENDPOINT + "/" + route
}

// login to server and start working session
func Login(endpoint string, username string, password string) error {
  input := map[string]interface{}{
    "username": username,
    "password": password,
  }
  data, err := Server(endpoint + "/login", Stringify(input))
  if err != nil { log.Fatal(err) }
  ACCESSTOKEN = String(data[TOKEN])
  if ACCESSTOKEN == ERROR {
    Deluser()
    Delzone()
    Delzadm()
    return fmt.Errorf("%s", String(data[TOKENTYPE]))
  } else {
    Setuser(String(data[USERATTR]))
    Setzone(String(data[ZONEATTR]))
    Setzadm(String(data[ZADMATTR]))
    r := String(data[USERATTR])
    fmt.Println(AUTHORIZED(), BACKCOLON, r)
    if True(SaveToken(ACCESSTOKEN)) { Error(ERRORSAVE()) }
    return nil
  }
}

// set/delete/get user name of working session
func Setuser(name string) {os.Setenv(SESSIONUSER, name)}
func Deluser() {os.Setenv(SESSIONUSER, "")}
func Getuser() string {return os.Getenv(SESSIONUSER)}

// set/delete/get client id of working session
func Setzone(id string) {os.Setenv(SESSIONZONE, id)}
func Delzone() {os.Setenv(SESSIONZONE, "")}
func Getzone() string {return os.Getenv(SESSIONZONE)}

// set/delete/get user admin privilege of working session
func Setzadm(status string) {os.Setenv(SESSIONZADM, status)}
func Delzadm() {os.Setenv(SESSIONZADM, "")}
func Getzadm() string {return os.Getenv(SESSIONZADM)}

// call server with json-string parameter using both methods
func Server(endpoint string, sparam string) (map[string]interface{}, error) {
  if APIACCESS == OAUTH2 {
    return ServerDo(endpoint, sparam, ACCESSTOKEN)
  } else {
    return ServerGet(endpoint, sparam)
  }
}

// call server with json-string parameter using stored token
func ServerDo(endpoint string, sparam string, access_token string) (map[string]interface{}, error) {
  req, err := http.NewRequest("GET", Url(endpoint, sparam), nil)
  if True(err) { return nil, err }
  req.Header.Set("Authorization", "Bearer " + access_token)
  resp, err := app_session.Do(req)
  if True(err) { return nil, err }
  defer resp.Body.Close()
  return Response(resp)
}

// call server with json-string parameter using cookie
func ServerGet(endpoint string, sparam string) (map[string]interface{}, error) {
  resp, err := app_session.Get(Url(endpoint, sparam))
  if True(err) { return nil, err }
  defer resp.Body.Close()
  return Response(resp)
}

// process server response
func Response(resp *http.Response) (map[string]interface{}, error) {
  body, err := ioutil.ReadAll(resp.Body)
  if True(err) { return nil, err }
  var data map[string]interface{}
  err = json.Unmarshal(body, &data)
  if True(err) { return nil, err }
  return data, nil
}

// parse url string for get request
func Url(endpoint string, sparam string) string {
  if sparam == "" { return endpoint }
  params := url.Values{}
  for key, value := range Parse(sparam) { params.Set(key, String(value)) }
  return fmt.Sprintf("%s?%s", endpoint, params.Encode())
}

func Cookie() {
  if APIACCESS == OAUTH2 {
    // use saved access token
    app_session = &http.Client {
      Transport: &http.Transport {
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
      },
    }
  } else {
    // use cookie to store session
    jar, err := cookiejar.New(nil)
    if err != nil { log.Fatal(err) }
    app_session = &http.Client{
      Jar: jar,
      Timeout: 10 * time.Second,
    }
  }
}

func SaveToken(token string) error {
  data, err := json.MarshalIndent(map[string]string{TOKEN: token}, "", "")
  if True(err) { return err }
  return ioutil.WriteFile(TOKENJSON, data, 0644)
}

func LoadToken() (string, error) {
  data, err := ioutil.ReadFile(TOKENJSON)
  if True(err) { return "", err }
  var token map[string]string
  err = json.Unmarshal(data, &token)
  if True(err) { return "", err }
  return Get(token, TOKEN), nil
}

// get user id by name
func Uid(name string) int {
  data := Result(DirectCall("select_user_id_by_name", Jname(name)))
  if False(data) { return -1 }
  return Integer(FromAny(data[0].(map[string]any))[IDATTR])
}

// get client id by name
func Zid(name string) int {
  data := Result(DirectCall("select_zone_id_by_name", Jname(name)))
  if False(data) { return -1 }
  return Integer(FromAny(data[0].(map[string]any))[IDATTR])
}

// update input["user"] from ["name"]
func User(input map[string]interface{}) {
  if Pull(input, USERATTR) == nil {
    input[USERATTR] = Uid(String(Pull(input, NAMEATTR)))
  }
}

// caution to ambiguous ["name"]
func Zone(input map[string]interface{}) {
  if Pull(input, ZONEATTR) == nil {
    input[ZONEATTR] = Zid(String(Pull(input, NAMEATTR)))
  }
}
