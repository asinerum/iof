// file io lib

package iof

import (
  "os"
  "log"
  "strings"
  "encoding/csv"
  "encoding/json"

  "gopkg.in/yaml.v3"
)

// load data from json file
func Json(path string) map[string]any {
  file, err := os.ReadFile(path)
  if err != nil { log.Fatal(err) }
  var result map[string]any
  err = json.Unmarshal(file, &result)
  if err != nil { log.Fatal(err) }
  return result
}

// load data from large csv file
func Csv(path string) [][]string {
  file, err := os.Open(path)
  if err != nil { log.Fatal(err) }
  defer file.Close()
  reader := csv.NewReader(file)
  records, err := reader.ReadAll()
  if err != nil { log.Fatal(err) }
  return records
}

// load data from csv text/string
func CSV(data string) [][]string {
  r := strings.NewReader(data)
  reader := csv.NewReader(r)
  records, err := reader.ReadAll()
  if err != nil { log.Fatal(err) }
  return records
}

// load app config from app.yaml
func App() map[string]interface{} {
  return Yaml(APPYAML)
}

// load menu.yaml for menu options
func Menu() map[string][]string {
  return YBS(MENUYAML)
}

// load struct.yaml into string slices
func Struct() map[string][]string {
  return YBS(STRUCTYAML)
}

// get struct heads from head.yaml
func Heads(id string) map[string]string {
  heads, ok := Head()[id]
  if !ok { log.Fatal(INVALIDSTRUCT()) }
  return heads
}

// get structs from struct.yaml
func Structing(id string) []string {
  structure, ok := Struct()[id]
  if !ok { log.Fatal(INVALIDSTRUCT()) }
  return structure
}

// load head.yaml into string map
func Head() map[string]map[string]string {
  return YBB(HEADYAML)
}

// get user-friendly headers for data output
func Header(heading map[string]string, structure []string) []string {
  heads := make([]string, len(structure))
  for index, val := range structure {
    if head, ok := heading[val]; ok {
      heads[index] = head
    } else {
      heads[index] = val
    }
  }
  return heads
}

// get headers from head.yaml for output
func Heading(id string) []string {
  structure := Structing(id)
  heading, ok := Head()[id]
  if !ok { return structure }
  return Header(heading, structure)
}

// load caption.yaml into string map
func Caption() map[string]map[string]string {
  return YBB(CAPTIONYAML)
}

// get captions for module <id>
func Captioner(id string) map[string]string {
  captions, ok := Caption()[id]
  if !ok { return nil }
  return captions
}

// get exact caption in module <id> using default value
func Captioning(id string, attr string, defval string) string {
  captions := Captioner(id)
  if captions == nil { return defval }
  caption, ok := captions[attr]
  if !ok { return defval }
  return caption
}

// load instruct.yaml into string slices
func Instruct() map[string][]string {
  return YBS(INSTRUCTYAML)
}

// load inscript.yaml into map
func Inscript() map[string]map[string]map[string]string {
  // need to convert [string] item to proper type
  return BYBB(SCRIPTYAML)
}

// load label.yaml for labels
func Label() map[string]string {
  return YB(LABELYAML)
}

// load text.yaml for texts
func Text() map[string]string {
  return YB(TEXTYAML)
}

// load print.yaml for labels
func Output() map[string]map[string]string {
  return YBB(OUTPUTYAML)
}

// load reform.yaml for output models
func Reform() map[string]map[string]string {
  return YBB(REFORMYAML)
}

// load return.yaml for output models
func Return() map[string]map[string]string {
  return YBB(RETURNYAML)
}

// load task.yaml for app tasks
func Task() map[string][]map[string]string {
  return YBBS(TASKYAML)
}

// load data.yaml for prebuilt data list
func Data() map[string][]map[string]string {
  return YBBS(DATAYAML)
}

// get all app top-lavel tasks
func TopTask() []map[string]string {
  return Task()["toptasks"]
}

// get all hidden tasks
func HidTask() []map[string]string {
  return Task()["hidtasks"]
}

// load yaml with map[string]string
func YB(path string) map[string]string {
  file := Yamlb(path)
  data := make(map[string]string)
  err := yaml.Unmarshal(file, &data)
  if err != nil { log.Fatal(err) }
  return data
}

// load yaml with map[string][]string
func YBS(path string) map[string][]string {
  file := Yamlb(path)
  data := make(map[string][]string)
  err := yaml.Unmarshal(file, &data)
  if err != nil { log.Fatal(err) }
  return data
}

// load yaml with map[string]map[string]string
func YBB(path string) map[string]map[string]string {
  file := Yamlb(path)
  data := make(map[string]map[string]string)
  err := yaml.Unmarshal(file, &data)
  if err != nil { log.Fatal(err) }
  return data
}

// load yaml with map[string][]map[string]string
func YBBS(path string) map[string][]map[string]string {
  file := Yamlb(path)
  data := make(map[string][]map[string]string)
  err := yaml.Unmarshal(file, &data)
  if err != nil { log.Fatal(err) }
  return data
}

// load yaml with map[string]map[string]map[string]string
func BYBB(path string) map[string]map[string]map[string]string {
  file := Yamlb(path)
  data := make(map[string]map[string]map[string]string)
  err := yaml.Unmarshal(file, &data)
  if err != nil { log.Fatal(err) }
  return data
}

// load structured data from yaml file
func Yaml(path string) map[string]interface{} {
  return YAML(Yamlb(path))
}

// load raw bytes from yaml file
func Yamlb(path string) []byte {
  file, err := os.ReadFile(path)
  if err != nil { log.Fatal(err) }
  return file
}

// load data from yaml bytes slice
func YAML(data []byte) map[string]interface{} {
  var config map[string]interface{}
  err := yaml.Unmarshal(data, &config)
  if err != nil { log.Fatal(err) }
  return config
}
