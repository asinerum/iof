// const and vars

package iof

import (
  "net/http"
)

type StrFunc func(name string) string // for data reform

type Standalone struct {} // container of functions

const COOKIE = "cookie"
const OAUTH2 = "oauth2"

const BACKCOLON = "\b:"
const BACKCOMMA = "\b,"
const BACKDOT = "\b."
const COLON = ":"
const COMMA = ","
const CR = '\n'
const DOT = "."
const HYPHEN = "-"
const SPACE = " "
const TAB = "\t"
const THREEDOT = "..."

const EXIT = "exit"
const QUIT = "quit"

const ERROR = "error"
const FOOTER = "footer"
const HEADER = "header"
const MESSAGE = "message"
const RESULT = "result"

const TOKEN = "access_token"
const TOKENTYPE = "token_type"

const IDATTR = "id"
const NAMEATTR = "name"
const USERATTR = "user"
const ZONEATTR = "zone"
const ZADMATTR = "zadmin"

const SESSIONZADM = "session_zadm"
const SESSIONUSER = "session_user"
const SESSIONZONE = "session_zone"

const PAGE = "page"
const LIMIT = "limit"
const OFFSET = "offset"

const DEFLIMIT = 50
const DEFOFFSET = 0

const APPYAML = "app.yaml"
const DATAYAML = "data.yaml"

const MENUYAML = "menu.yaml"
const LABELYAML = "label.yaml"

const STRUCTYAML = "outstruct.yaml"
const HEADYAML = "outhead.yaml"
const CAPTIONYAML = "outcaption.yaml"

const REFORMYAML = "reform.yaml"
const RETURNYAML = "return.yaml"

const INSTRUCTYAML = "instruct.yaml"
const SCRIPTYAML = "inscript.yaml"

const TASKYAML = "task.yaml"
const TEXTYAML = "text.yaml"

const OUTPUTYAML = "print.yaml"
const TOKENJSON = "token.json"

var DATA map[string][]map[string]string
var OUTPUTS map[string]map[string]string
var REFORMS map[string]map[string]string
var RETURNS map[string]map[string]string
var HIDTASKS []map[string]string
var TOPTASKS []map[string]string
var LABELS map[string]string
var TEXTS map[string]string

var StrFunctions map[string]StrFunc
var app_session *http.Client

var SERVERENDPOINT string
var ACCESSTOKEN string
var APIACCESS string

var PROMPT string
var YES string
var NO string
var CANCEL string
